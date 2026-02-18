package app_filter

import (
	"src/internal/platform/proc_sensing"
)

// Filter provides methods to determine if a process should be tracked or logged.
type Filter interface {
	// ShouldTrack returns true if the process is a user application that should be monitored.
	ShouldTrack(exePath string, proc *proc_sensing.ProcessInfo) bool

	// ShouldExclude returns true if the process is a system component that should be ignored.
	ShouldExclude(exePath string, proc *proc_sensing.ProcessInfo) bool
}
