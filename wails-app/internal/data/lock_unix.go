//go:build !windows

package data

// platformLock is a dummy implementation for non-Windows platforms.
func platformLock(path string) error {
	return nil
}
