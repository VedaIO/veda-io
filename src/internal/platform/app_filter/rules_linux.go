//go:build linux

package app_filter

import (
	"src/internal/platform/proc_sensing"
	"strings"
)

// ShouldExclude returns true if the process should be ignored (Stub for non-Windows).
func ShouldExclude(exePath string, proc *proc_sensing.ProcessInfo) bool {
	exePathLower := strings.ToLower(exePath)

	// Rule 0: Never track ProcGuard itself
	if strings.Contains(exePathLower, "procguard") {
		return true
	}

	return false
}

// ShouldTrack returns true if the process should be tracked (Stub for non-Windows).
func ShouldTrack(exePath string, proc *proc_sensing.ProcessInfo) bool {
	// Simple heuristic for Linux: everything not excluded is tracked
	return true
}
