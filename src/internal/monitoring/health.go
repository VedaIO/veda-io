package monitoring

import (
	"time"
)

// HealthStatus represents the current health state of the monitoring manager.
// It is used for diagnostics and monitoring system stability.
type HealthStatus struct {
	// IsHealthy indicates whether the monitoring loop is currently running.
	IsHealthy bool
	// LastTickTime is the timestamp of the last successful process snapshot.
	LastTickTime time.Time
	// ConsecutiveFailures is the number of consecutive failures since the last success.
	ConsecutiveFailures int32
	// SubscriberCount is the number of registered subscribers.
	SubscriberCount int
}

// HealthCheck returns the current health status of the monitoring manager.
// This can be called externally to monitor the system's health.
func (m *MonitoringManager) HealthCheck() HealthStatus {
	lastTick := m.lastTickTime.Load().(time.Time)
	return HealthStatus{
		IsHealthy:           m.isRunning.Load(),
		LastTickTime:        lastTick,
		ConsecutiveFailures: m.consecutiveFailures.Load(),
		SubscriberCount:     len(m.subscribers),
	}
}
