//go:build linux

// Package icon provides executable icon extraction functionality.
package icon

// GetAppIconAsBase64 extracts the first icon from an executable
// and returns it as a base64-encoded PNG string.
// Returns empty string on Linux (not yet implemented).
func GetAppIconAsBase64(exePath string) (string, error) {
	// TODO: Implement using freedesktop icon theme
	return "", nil
}
