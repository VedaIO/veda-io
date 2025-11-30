//go:build !windows

package app

// GetProcessIntegrityLevel is a dummy implementation for non-Windows platforms.
func GetProcessIntegrityLevel(pid uint32) (uint32, error) {
	return 0, nil
}
