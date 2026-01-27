package screentime

import (
	"database/sql"
	"time"
	"wails-app/internal/data/logger"
	"wails-app/internal/data/write"
	platformScreentime "wails-app/internal/platform/screentime"

	"github.com/shirou/gopsutil/v3/process"
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
			ExeCache:      make(map[uint32]CachedProcInfo),
		}
		ticker := time.NewTicker(screenTimeCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				trackForegroundWindow(appLogger, state)
			case <-resetScreenTimeCh:
				appLogger.Printf("[Screentime] Reset signal received. Clearing in-memory state.")
				state.LastExePath = ""
				state.LastTitle = ""
				state.PendingDuration = 0
				state.ExeCache = make(map[uint32]CachedProcInfo)
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

	exePath := ""

	// Get process object to check creation time for validation
	proc, err := process.NewProcess(int32(info.PID))
	if err == nil {
		createTime, err := proc.CreateTime()
		if err == nil {
			// Check cache
			cached, ok := state.ExeCache[info.PID]
			if ok && cached.CreationTime == createTime {
				// Cache hit and validated!
				exePath = cached.ExePath
			} else {
				// Cache miss or obsolete PID. Resolve path.
				path, err := proc.Exe()
				if err == nil {
					exePath = path
					state.ExeCache[info.PID] = CachedProcInfo{
						ExePath:      path,
						CreationTime: createTime,
					}
				}
			}
		}
	}

	if exePath == "" {
		// Fallback if we couldn't get creation time or process info (rare race condition or permission issue)
		return
	}

	// Filter out applications that should not be tracked (e.g., system services).
	if !ShouldTrackApp(exePath, proc) {
		return
	}

	// Check if the user is still in the same window as the last check.
	if exePath == state.LastExePath && info.Title == state.LastTitle {
		// Same window: increment the memory buffer. We don't write to DB yet.
		state.PendingDuration++
	} else {
		// Window changed: we must flush the accumulated time for the *previous* window.
		if state.PendingDuration > 0 {
			flushScreenTime(appLogger, state.LastExePath, state.LastTitle, state.PendingDuration)
		}

		// Insert a new record for the *new* window session.
		// We insert with 1 second duration to establish the record.
		now := time.Now().Unix()
		appLogger.Printf("[Screentime] New window: %s (%s)", exePath, info.Title)
		write.EnqueueWrite(`
			INSERT INTO screen_time (executable_path, window_title, timestamp, duration_seconds)
			VALUES (?, ?, ?, 1)
		`, exePath, info.Title, now)

		// Update our state to reflect the new active window.
		state.LastExePath = exePath
		state.LastTitle = info.Title
		state.PendingDuration = 0 // Duration is 0 because we just inserted 1s in the DB.
	}

	// Periodically flush the buffer to the DB, even if the window hasn't changed.
	// This ensures the UI shows relatively up-to-date data during long sessions in one app.
	if time.Since(state.LastFlushTime) >= dbFlushInterval {
		if state.PendingDuration > 0 {
			flushScreenTime(appLogger, state.LastExePath, state.LastTitle, state.PendingDuration)
			state.PendingDuration = 0 // Reset buffer after flush.
		}
		state.LastFlushTime = time.Now()
	}
}
