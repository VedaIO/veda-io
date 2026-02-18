package app

import (
	"fmt"
	"src/internal/data/logger"
	"src/internal/data/repository"
	"src/internal/platform/app_filter"
	"src/internal/platform/proc_sensing"
	"strings"
	"sync"
	"time"
)

const processCheckInterval = 2 * time.Second

// loggerState encapsulates the in-memory state of the process monitor.
type loggerState struct {
	runningProcs     map[string]string // UniqueKey (PID-StartTime) -> lowercase process name
	runningAppCounts map[string]int    // lowercase process name -> instance count
	sync.Mutex
}

var resetLoggerCh = make(chan struct{}, 1)

// ResetLoggedApps clears the in-memory cache of logged applications.
func ResetLoggedApps() {
	resetLoggerCh <- struct{}{}
}

// StartProcessEventLogger starts a long-running goroutine that monitors process creation and termination events.
func StartProcessEventLogger(appLogger logger.Logger, repo *repository.AppRepository) {
	state := &loggerState{
		runningProcs:     make(map[string]string),
		runningAppCounts: make(map[string]int),
	}

	go func() {
		initializeRunningProcs(state, repo)

		ticker := time.NewTicker(processCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				procs, err := proc_sensing.GetAllProcesses()
				if err != nil {
					appLogger.Printf("Failed to get processes: %v", err)
					continue
				}

				currentKeys := make(map[string]bool)
				for _, p := range procs {
					currentKeys[p.UniqueKey()] = true
				}

				logEndedProcesses(state, currentKeys, repo)
				logNewProcesses(state, appLogger, procs, repo)
			case <-resetLoggerCh:
				appLogger.Printf("[Logger] Reset signal received. Clearing in-memory state.")
				state.Lock()
				state.runningProcs = make(map[string]string)
				state.runningAppCounts = make(map[string]int)
				state.Unlock()
			}
		}
	}()
}

func logEndedProcesses(state *loggerState, currentKeys map[string]bool, repo *repository.AppRepository) {
	state.Lock()
	defer state.Unlock()

	for key, nameLower := range state.runningProcs {
		if !currentKeys[key] {
			// process_instance_key (PID-StartTimeNano) is used to uniquely identify the session row.
			repo.CloseAppEvent(key, time.Now().Unix())

			delete(state.runningProcs, key)
			state.runningAppCounts[nameLower]--
			if state.runningAppCounts[nameLower] <= 0 {
				delete(state.runningAppCounts, nameLower)
			}
		}
	}
}

func logNewProcesses(state *loggerState, appLogger logger.Logger, procs []proc_sensing.ProcessInfo, repo *repository.AppRepository) {
	state.Lock()
	defer state.Unlock()

	for _, p := range procs {
		key := p.UniqueKey()
		if _, exists := state.runningProcs[key]; exists {
			continue
		}

		name := p.Name
		if name == "" {
			continue
		}
		nameLower := strings.ToLower(name)

		exePath := p.ExePath
		if exePath == "" {
			continue
		}

		// Rule 1: Platform-specific system exclusion
		if app_filter.ShouldExclude(exePath, &p) {
			state.runningProcs[key] = nameLower
			continue
		}

		// Rule 2: Deduplication (Reference Counting)
		isAlreadyLogged := state.runningAppCounts[nameLower] > 0
		if isAlreadyLogged {
			state.runningProcs[key] = nameLower
			state.runningAppCounts[nameLower]++
			continue
		}

		// Rule 3: Must be a trackable user application
		if !app_filter.ShouldTrack(exePath, &p) {
			continue
		}

		// Success: Log it
		parentName := fmt.Sprintf("PID: %d", p.ParentPID)

		// Store a standard Unix timestamp for display and the high-precision UniqueKey for process identity.
		repo.LogAppEvent(name, p.PID, parentName, exePath, time.Now().Unix(), key)

		state.runningProcs[key] = nameLower
		state.runningAppCounts[nameLower]++
	}
}

func initializeRunningProcs(state *loggerState, repo *repository.AppRepository) {
	// Restore state primarily using the precise key via the repository
	activeSessions, err := repo.GetActiveSessions()
	if err != nil {
		return
	}

	state.Lock()
	defer state.Unlock()

	// Pre-fetch current processes for validation
	procs, _ := proc_sensing.GetAllProcesses()
	currentKeys := make(map[string]bool)
	for _, p := range procs {
		currentKeys[p.UniqueKey()] = true
	}

	for _, s := range activeSessions {
		if currentKeys[s.Key] {
			nameLower := strings.ToLower(s.Name)
			state.runningProcs[s.Key] = nameLower
			state.runningAppCounts[nameLower]++
		} else {
			// Process no longer running (or PID recycled/different instance) - Close it!
			repo.CloseAppEvent(s.Key, time.Now().Unix())
		}
	}
}
