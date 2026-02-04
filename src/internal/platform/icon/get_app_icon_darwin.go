//go:build darwin

// Package icon provides executable icon extraction functionality.
package icon

// GetAppIconAsBase64 extracts the first icon from an executable
// and returns it as a base64-encoded PNG string.
// Returns empty string on Darwin (not yet implemented).
func GetAppIconAsBase64(exePath string) (string, error) {
	// TODO: Implement using NSWorkspace iconForFile
	return "", nil
}
