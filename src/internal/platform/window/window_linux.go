//go:build linux

// Package window provides functionality to check window visibility.
package window

// HasVisibleWindow checks if a process with the given PID has a visible window.
// Returns true on Linux as a safe default.
func HasVisibleWindow(pid uint32) bool {
	return true
}
