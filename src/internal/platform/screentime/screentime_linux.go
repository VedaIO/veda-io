//go:build linux

// Package screentime provides foreground window tracking functionality.
package screentime

// WindowInfo contains information about the currently focused window
type WindowInfo struct {
	PID   uint32
	Title string
}

// GetActiveWindowInfo retrieves the PID and title of the foreground window.
// Returns nil on Linux (not yet implemented).
func GetActiveWindowInfo() *WindowInfo {
	// TODO: Implement using X11 or Wayland APIs
	return nil
}
