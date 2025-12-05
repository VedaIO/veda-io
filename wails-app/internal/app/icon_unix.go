//go:build !windows

package app

// GetAppIconAsBase64 is a dummy implementation for non-Windows platforms.
func GetAppIconAsBase64(exePath string) (string, error) {
	return "", nil
}
