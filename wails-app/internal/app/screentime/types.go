package screentime

import (
	"time"
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
	LastExePath string
	// lastTitle is the title of the previously detected foreground window.
	LastTitle string
	// pendingDuration is the number of seconds the current window has been active since the last DB flush.
	PendingDuration int
	// lastFlushTime is the timestamp of the last successful database flush.
	LastFlushTime time.Time
	// exeCache maps Process IDs (PID) to their cached info (path + validation data).
	ExeCache map[uint32]CachedProcInfo
}
