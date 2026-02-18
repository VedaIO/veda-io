package screentime

import (
	"src/internal/data/logger"
	"src/internal/data/repository"
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
func StartScreenTimeMonitor(appLogger logger.Logger, apps *repository.AppRepository, web *repository.WebRepository) {
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
				trackForegroundWindow(appLogger, state, apps, web)
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
func trackForegroundWindow(appLogger logger.Logger, state *ScreenTimeState, apps *repository.AppRepository, web *repository.WebRepository) {
	// Retrieve the active window information from the platform-specific implementation.
	info := platformScreentime.GetActiveWindowInfo()
	if info == nil || info.PID == 0 {
		return
	}

	// Use targeted sensing for the specific foreground PID to minimize system overhead.
	activeProc, err := proc_sensing.GetProcessByPID(info.PID)
	if err != nil {
		appLogger.Printf("[Screentime] GetProcessByPID failed for PID %d: %v", info.PID, err)
		return
	}

	uniqueKey := activeProc.UniqueKey()
	exePath := activeProc.ExePath

	// Filter out applications that should not be tracked (e.g., system services).
	if app_filter.ShouldExclude(exePath, &activeProc) {
		appLogger.Printf("[Screentime] Excluded by filter: %s", exePath)
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
			appLogger.Printf("[Screentime] Flushing %d seconds for %s", state.PendingDuration, state.LastExePath)
			apps.UpdateScreenTime(state.LastExePath, state.PendingDuration)
		}

		// Update existing record for this app if it was recently active, otherwise create a new one.
		appLogger.Printf("[Screentime] Focus shifted to: %s", exePath)

		// Attempt to update a record if it was active in the last 5 minutes.
		apps.EnsureActiveScreenTimeRecord(exePath)

		// Update our state to reflect the new active window.
		state.LastUniqueKey = uniqueKey
		state.LastExePath = exePath
		state.PendingDuration = 0
	}

	// Periodically flush the buffer to the DB, even if the window hasn't changed.
	// This ensures the UI shows relatively up-to-date data during long sessions in one app.
	if time.Since(state.LastFlushTime) >= dbFlushInterval {
		if state.PendingDuration > 0 {
			appLogger.Printf("[Screentime] Periodic flush: %d seconds for %s", state.PendingDuration, state.LastExePath)
			apps.UpdateScreenTime(state.LastExePath, state.PendingDuration)
			state.PendingDuration = 0 // Reset buffer after flush.
		}
		state.LastFlushTime = time.Now()
	}
}
