//go:build linux

package uninstall

import "fmt"

// SelfDestruct is a dummy implementation for non-Windows platforms.
func SelfDestruct(appName string) error {
	return fmt.Errorf("SelfDestruct is only supported on Windows")
}
