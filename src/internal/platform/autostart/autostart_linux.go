//go:build linux

package autostart

// EnsureAutostart is a dummy implementation for non-Windows platforms.
func EnsureAutostart() (string, error) {
	return "", nil
}

// RemoveAutostart is a dummy implementation for non-Windows platforms.
func RemoveAutostart() error {
	return nil
}
