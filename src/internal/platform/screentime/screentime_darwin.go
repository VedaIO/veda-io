//go:build darwin

// Package screentime provides foreground window tracking functionality.
package screentime

// WindowInfo contains information about the currently focused window
type WindowInfo struct {
	PID   uint32
	Title string
}

// GetActiveWindowInfo retrieves the PID and title of the foreground window.
// Returns nil on Darwin (not yet implemented).
func GetActiveWindowInfo() *WindowInfo {
	// TODO: Implement using NSWorkspace / CGWindowListCopyWindowInfo
	return nil
}
