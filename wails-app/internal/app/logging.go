package app

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
	"wails-app/internal/data/logger"
	"wails-app/internal/data/write"
	"wails-app/internal/platform/app_filter"
	"wails-app/internal/platform/proc_sensing"
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
func StartProcessEventLogger(appLogger logger.Logger, db *sql.DB) {
	state := &loggerState{
		runningProcs:     make(map[string]string),
		runningAppCounts: make(map[string]int),
	}

	go func() {
		initializeRunningProcs(state, db)

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

				logEndedProcesses(state, currentKeys)
				logNewProcesses(state, appLogger, procs)
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

// ... (StartProcessEventLogger remains same)

func logEndedProcesses(state *loggerState, currentKeys map[string]bool) {
	state.Lock()
	defer state.Unlock()

	for key, nameLower := range state.runningProcs {
		if !currentKeys[key] {
			// We use the UniqueKey (PID-StartTimeNano) stored in process_instance_key to identify the row.
			// Note: For backward compatibility with very old rows (no key), we might need a fallback,
			// but for this fix we assume rows created since this update have the key.
			write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE process_instance_key = ? AND end_time IS NULL",
				time.Now().Unix(), key)

			delete(state.runningProcs, key)
			state.runningAppCounts[nameLower]--
			if state.runningAppCounts[nameLower] <= 0 {
				delete(state.runningAppCounts, nameLower)
			}
		}
	}
}

func logNewProcesses(state *loggerState, appLogger logger.Logger, procs []proc_sensing.ProcessInfo) {
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

		// KEY FIX: We store time.Now().Unix() in start_time for UI compatibility (Seconds).
		// We store p.UniqueKey() in process_instance_key for precise tracking.
		write.EnqueueWrite("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time, process_instance_key) VALUES (?, ?, ?, ?, ?, ?)",
			name, p.PID, parentName, exePath, time.Now().Unix(), key)

		state.runningProcs[key] = nameLower
		state.runningAppCounts[nameLower]++
	}
}

func initializeRunningProcs(state *loggerState, db *sql.DB) {
	// Restore state primarily using the precise key
	rows, err := db.Query("SELECT pid, process_name, process_instance_key FROM app_events WHERE end_time IS NULL AND process_instance_key IS NOT NULL")
	if err != nil {
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close rows: %v", err)
		}
	}()

	state.Lock()
	defer state.Unlock()

	// Pre-fetch current processes for validation
	procs, _ := proc_sensing.GetAllProcesses()
	currentKeys := make(map[string]bool)
	for _, p := range procs {
		currentKeys[p.UniqueKey()] = true
	}

	for rows.Next() {
		var pid int32
		var name string
		var key string
		if err := rows.Scan(&pid, &name, &key); err == nil {
			if currentKeys[key] {
				nameLower := strings.ToLower(name)
				state.runningProcs[key] = nameLower
				state.runningAppCounts[nameLower]++
			} else {
				// Process no longer running (or PID recycled/different instance) - Close it!
				write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE process_instance_key = ? AND end_time IS NULL",
					time.Now().Unix(), key)
			}
		}
	}
}
