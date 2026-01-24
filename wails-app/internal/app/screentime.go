package app

import (
	"database/sql"
	"time"
	"wails-app/internal/data"
	"wails-app/internal/data/write"
	"wails-app/internal/platform/screentime"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	// screenTimeCheckInterval determines how often we check the foreground window.
	screenTimeCheckInterval = 1 * time.Second
	// dbFlushInterval determines how often we flush buffered screen time data to the database.
	dbFlushInterval = 10 * time.Second
)

// CachedProcInfo stores the executable path and creation time of a process.
// We use the Creation Time to validate that a PID hasn't been reused.
type CachedProcInfo struct {
	ExePath      string
	CreationTime int64 // Unix timestamp in milliseconds
}

// ScreenTimeState maintains the state of the screen time monitoring loop.
// It buffers database writes and caches process information to improve performance.
type ScreenTimeState struct {
	// lastExePath is the executable path of the previously detected foreground window.
	lastExePath string
	// lastTitle is the title of the previously detected foreground window.
	lastTitle string
	// pendingDuration is the number of seconds the current window has been active since the last DB flush.
	pendingDuration int
	// lastFlushTime is the timestamp of the last successful database flush.
	lastFlushTime time.Time
	// exeCache maps Process IDs (PID) to their cached info (path + validation data).
	exeCache map[uint32]CachedProcInfo
}

var resetScreenTimeCh = make(chan struct{}, 1)

// ResetScreenTime clears the in-memory screen time state and process cache.
func ResetScreenTime() {
	resetScreenTimeCh <- struct{}{}
}

// StartScreenTimeMonitor initializes and starts the background goroutine for tracking screen time.
// It uses a ticker to poll the foreground window at regular intervals and buffers updates
// to reduce database I/O.
func StartScreenTimeMonitor(appLogger data.Logger, db *sql.DB) {
	go func() {
		state := &ScreenTimeState{
			lastFlushTime: time.Now(),
			exeCache:      make(map[uint32]CachedProcInfo),
		}
		ticker := time.NewTicker(screenTimeCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				trackForegroundWindow(appLogger, state)
			case <-resetScreenTimeCh:
				appLogger.Printf("[Screentime] Reset signal received. Clearing in-memory state.")
				state.lastExePath = ""
				state.lastTitle = ""
				state.pendingDuration = 0
				state.exeCache = make(map[uint32]CachedProcInfo)
			}
		}
	}()
}

// trackForegroundWindow performs a single check of the active window and updates the state.
func trackForegroundWindow(appLogger data.Logger, state *ScreenTimeState) {
	// Retrieve the active window information from the platform-specific implementation.
	info := screentime.GetActiveWindowInfo()
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
			cached, ok := state.exeCache[info.PID]
			if ok && cached.CreationTime == createTime {
				// Cache hit and validated!
				exePath = cached.ExePath
			} else {
				// Cache miss or obsolete PID. Resolve path.
				path, err := proc.Exe()
				if err == nil {
					exePath = path
					state.exeCache[info.PID] = CachedProcInfo{
						ExePath:      path,
						CreationTime: createTime,
					}
					// appLogger.Printf("[Perf] Cached PID %d -> %s", info.PID, path) // Too verbose?
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
	if exePath == state.lastExePath && info.Title == state.lastTitle {
		// Same window: increment the memory buffer. We don't write to DB yet.
		state.pendingDuration++
	} else {
		// Window changed: we must flush the accumulated time for the *previous* window.
		if state.pendingDuration > 0 {
			flushScreenTime(appLogger, state.lastExePath, state.lastTitle, state.pendingDuration)
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
		state.lastExePath = exePath
		state.lastTitle = info.Title
		state.pendingDuration = 0 // Duration is 0 because we just inserted 1s in the DB.
	}

	// Periodically flush the buffer to the DB, even if the window hasn't changed.
	// This ensures the UI shows relatively up-to-date data during long sessions in one app.
	if time.Since(state.lastFlushTime) >= dbFlushInterval {
		if state.pendingDuration > 0 {
			flushScreenTime(appLogger, state.lastExePath, state.lastTitle, state.pendingDuration)
			state.pendingDuration = 0 // Reset buffer after flush.
		}
		state.lastFlushTime = time.Now()
	}
}

// flushScreenTime writes the buffered duration to the database.
// It updates the most recent record for the given app and title, adding the buffered duration.
func flushScreenTime(logger data.Logger, exePath, title string, duration int) {
	write.EnqueueWrite(`
		UPDATE screen_time 
		SET duration_seconds = duration_seconds + ?
		WHERE id = (
			SELECT id FROM screen_time 
			WHERE executable_path = ? AND window_title = ?
			ORDER BY timestamp DESC LIMIT 1
		)
	`, duration, exePath, title)
}
