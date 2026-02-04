package screentime

import (
	"time"
)

// CachedProcInfo stores the executable path and creation time of a process.
type CachedProcInfo struct {
	ExePath       string
	StartTimeNano uint64
}

// ScreenTimeState maintains the state of the screen time monitoring loop.
// It buffers database writes and caches process information to improve performance.
type ScreenTimeState struct {
	// LastUniqueKey is the unique identifier (PID-StartTime) of the previously detected process.
	LastUniqueKey string
	// lastExePath is the executable path of the previously detected foreground window.
	LastExePath string
	// pendingDuration is the number of seconds the current window has been active since the last DB flush.
	PendingDuration int
	// lastFlushTime is the timestamp of the last successful database flush.
	LastFlushTime time.Time
	// ExeCache maps unique process keys (PID-StartTime) to their cached info (path + validation data).
	ExeCache map[string]CachedProcInfo
}
