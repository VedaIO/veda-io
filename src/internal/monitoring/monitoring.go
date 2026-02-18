package monitoring

import (
	"sync"
	"sync/atomic"
	"time"

	"src/internal/data/logger"
	"src/internal/platform/proc_sensing"
)

// DefaultPollingInterval is the default interval between process snapshots.
const DefaultPollingInterval = 2 * time.Second

// MonitoringManager is the core component that orchestrates process monitoring.
// It captures process snapshots at regular intervals and distributes them to registered subscribers.
// It also handles self-healing through panic recovery and automatic restart.
type MonitoringManager struct {
	logger              logger.Logger
	subscribers         []ProcessSubscriber
	pollingInterval     time.Duration
	stopCh              chan struct{}
	wg                  sync.WaitGroup
	processSnapshotCh   chan ProcessSnapshot
	resetCh             chan struct{}
	isRunning           atomic.Bool
	lastTickTime        atomic.Value
	restartDelay        time.Duration
	restartMaxRetries   int
	consecutiveFailures atomic.Int32
}

// globalManager is the singleton instance used for global reset operations.
var globalManager *MonitoringManager

// SetGlobalManager sets the global manager instance for external reset operations.
func SetGlobalManager(manager *MonitoringManager) {
	globalManager = manager
}

// ResetGlobalManager sends a reset signal to the global manager if one is set.
func ResetGlobalManager() {
	if globalManager != nil {
		globalManager.Reset()
	}
}

// NewMonitoringManager creates a new MonitoringManager with the specified logger and polling interval.
// If pollingInterval is zero or negative, DefaultPollingInterval is used.
func NewMonitoringManager(appLogger logger.Logger, pollingInterval time.Duration) *MonitoringManager {
	if pollingInterval <= 0 {
		pollingInterval = DefaultPollingInterval
	}
	m := &MonitoringManager{
		logger:            appLogger,
		subscribers:       make([]ProcessSubscriber, 0),
		pollingInterval:   pollingInterval,
		stopCh:            make(chan struct{}),
		processSnapshotCh: make(chan ProcessSnapshot, 1),
		resetCh:           make(chan struct{}, 1),
		restartDelay:      5 * time.Second,
		restartMaxRetries: 3,
	}
	m.lastTickTime.Store(time.Time{})
	return m
}

// RegisterSubscriber adds a subscriber to the list of components that receive process snapshots.
func (m *MonitoringManager) RegisterSubscriber(subscriber ProcessSubscriber) {
	m.subscribers = append(m.subscribers, subscriber)
}

// Start begins the monitoring loop in a new goroutine.
// It is safe to call Start multiple times - subsequent calls are ignored if already running.
func (m *MonitoringManager) Start() {
	if m.isRunning.Load() {
		m.logger.Printf("[MonitoringManager] Already running, skipping start")
		return
	}
	m.isRunning.Store(true)
	m.wg.Add(1)
	go m.runEventLoopWithRecovery()
}

// Stop gracefully stops the monitoring loop.
// It waits for the current iteration to complete before returning.
func (m *MonitoringManager) Stop() {
	if !m.isRunning.Load() {
		return
	}
	m.isRunning.Store(false)
	close(m.stopCh)
	m.wg.Wait()
}

// runEventLoopWithRecovery wraps the main event loop with panic recovery.
// If a panic occurs, it triggers the failure handling mechanism to attempt a restart.
func (m *MonitoringManager) runEventLoopWithRecovery() {
	defer func() {
		m.wg.Done()
		if r := recover(); r != nil {
			m.logger.Printf("[MonitoringManager] PANIC RECOVERY: %v", r)
			m.handleFailure()
		}
	}()
	m.runEventLoop()
}

// handleFailure manages the failure recovery process.
// It increments the failure counter, checks if max retries are exceeded,
// and schedules a restart if appropriate.
func (m *MonitoringManager) handleFailure() {
	m.isRunning.Store(false)
	m.consecutiveFailures.Add(1)
	failures := m.consecutiveFailures.Load()

	if failures >= int32(m.restartMaxRetries) {
		m.logger.Printf("[MonitoringManager restart] Max retries (%d) reached, giving up", m.restartMaxRetries)
		return
	}

	m.logger.Printf("[MonitoringManager] Restarting in %v (attempt %d/%d)",
		m.restartDelay, failures+1, m.restartMaxRetries)

	time.Sleep(m.restartDelay)

	m.logger.Printf("[MonitoringManager] Restarting monitoring loop...")
	m.Start()
}

// Reset sends a reset signal to all subscribers that implement ResettableSubscriber.
// This is typically called when clearing application history.
func (m *MonitoringManager) Reset() {
	m.logger.Printf("[MonitoringManager] Reset signal received")
	for _, subscriber := range m.subscribers {
		if resettable, ok := subscriber.(ResettableSubscriber); ok {
			resettable.Reset()
		}
	}
}

// runEventLoop is the main event loop that polls for process snapshots.
// It runs until the stop channel is closed.
func (m *MonitoringManager) runEventLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.pollingInterval)
	defer ticker.Stop()

	m.logger.Printf("[MonitoringManager] Started with polling interval: %v", m.pollingInterval)

	for {
		select {
		case <-m.stopCh:
			m.logger.Printf("[MonitoringManager] Stopping")
			return
		case snapshot := <-m.processSnapshotCh:
			m.notifySubscribers(snapshot)
		case <-m.resetCh:
			m.Reset()
		case <-ticker.C:
			m.captureAndNotify()
		}
	}
}

// captureAndNotify captures a process snapshot and distributes it to all subscribers.
// It uses the cached process list to reduce system overhead.
func (m *MonitoringManager) captureAndNotify() {
	procs, err := proc_sensing.GetAllProcessesCached()
	if err != nil {
		m.logger.Printf("[MonitoringManager] Failed to capture process snapshot: %v", err)
		return
	}

	m.lastTickTime.Store(time.Now())
	m.consecutiveFailures.Store(0)

	snapshot := ProcessSnapshot{
		Processes: procs,
		Timestamp: time.Now(),
	}

	m.notifySubscribers(snapshot)
}

// notifySubscribers distributes a process snapshot to all registered subscribers.
// Each subscriber is called in a protected wrapper that recovers from panics.
func (m *MonitoringManager) notifySubscribers(snapshot ProcessSnapshot) {
	for _, subscriber := range m.subscribers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					m.logger.Printf("[MonitoringManager] Subscriber %s panicked: %v", subscriber.Name(), r)
				}
			}()
			subscriber.OnProcessesChanged(snapshot)
		}()
	}
}
