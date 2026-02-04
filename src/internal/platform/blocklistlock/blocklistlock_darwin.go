//go:build darwin

package blocklistlock

// PlatformLock is a dummy implementation for non-Windows platforms.
func PlatformLock(path string) error {
	return nil
}
