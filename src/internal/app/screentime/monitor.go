package screentime

import (
	"database/sql"
	"src/internal/data/logger"
	"src/internal/data/write"
	"src/internal/platform/app_filter"
	"src/internal/platform/proc_sensing"
	platformScreentime "src/internal/platform/screentime"
	"time"
)

const (
	// screenTimeCheckInterval determines how often we check the foreground window.
	screenTimeCheckInterval = 1 * time.Second
	// dbFlushInterval determines how often we flush buffered screen time data to the database.
	dbFlushInterval = 10 * time.Second
)

var resetScreenTimeCh = make(chan struct{}, 1)

// ResetScreenTime clears the in-memory screen time state and process cache.
func ResetScreenTime() {
	resetScreenTimeCh <- struct{}{}
}

// StartScreenTimeMonitor initializes and starts the background goroutine for tracking screen time.
// It uses a ticker to poll the foreground window at regular intervals and buffers updates
// to reduce database I/O.
func StartScreenTimeMonitor(appLogger logger.Logger, db *sql.DB) {
	go func() {
		state := &ScreenTimeState{
			LastFlushTime: time.Now(),
			ExeCache:      make(map[string]CachedProcInfo),
		}
		ticker := time.NewTicker(screenTimeCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				trackForegroundWindow(appLogger, state)
			case <-resetScreenTimeCh:
				appLogger.Printf("[Screentime] Reset signal received. Clearing in-memory state.")
				state.LastUniqueKey = ""
				state.LastExePath = ""
				state.PendingDuration = 0
				state.ExeCache = make(map[string]CachedProcInfo)
			}
		}
	}()
}

// trackForegroundWindow performs a single check of the active window and updates the state.
func trackForegroundWindow(appLogger logger.Logger, state *ScreenTimeState) {
	// Retrieve the active window information from the platform-specific implementation.
	info := platformScreentime.GetActiveWindowInfo()
	if info == nil || info.PID == 0 {
		return
	}

	// Use targeted sensing for the specific foreground PID to minimize system overhead.
	activeProc, err := proc_sensing.GetProcessByPID(info.PID)
	if err != nil {
		return
	}

	uniqueKey := activeProc.UniqueKey()
	exePath := activeProc.ExePath

	// Filter out applications that should not be tracked (e.g., system services).
	if app_filter.ShouldExclude(exePath, &activeProc) {
		return
	}

	// track based on unique process instance identity.
	// We aggregate by executable path to avoid redundant entries when window titles change.
	if uniqueKey == state.LastUniqueKey {
		// Same app: increment the memory buffer. We don't write to DB yet.
		state.PendingDuration++
	} else {
		// App changed: flush the accumulated time for the *previous* app.
		if state.PendingDuration > 0 {
			flushScreenTime(appLogger, state.LastExePath, state.PendingDuration)
		}

		// Update existing record for this app if it was recently active, otherwise create a new one.
		now := time.Now().Unix()
		appLogger.Printf("[Screentime] Focus shifted to: %s", exePath)

		// Attempt to update a record if it was active in the last 5 minutes.
		write.EnqueueWrite(`
			INSERT INTO screen_time (executable_path, timestamp, duration_seconds)
			SELECT ?, ?, 1
			WHERE NOT EXISTS (
				SELECT 1 FROM screen_time 
				WHERE executable_path = ? AND timestamp > ?
				ORDER BY timestamp DESC LIMIT 1
			)
		`, exePath, now, exePath, now-300)

		// Update our state to reflect the new active window.
		state.LastUniqueKey = uniqueKey
		state.LastExePath = exePath
		state.PendingDuration = 0
	}

	// Periodically flush the buffer to the DB, even if the window hasn't changed.
	// This ensures the UI shows relatively up-to-date data during long sessions in one app.
	if time.Since(state.LastFlushTime) >= dbFlushInterval {
		if state.PendingDuration > 0 {
			flushScreenTime(appLogger, state.LastExePath, state.PendingDuration)
			state.PendingDuration = 0 // Reset buffer after flush.
		}
		state.LastFlushTime = time.Now()
	}
}
