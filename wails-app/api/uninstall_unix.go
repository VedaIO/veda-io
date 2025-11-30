//go:build !windows

package api

import (
	"fmt"
)

// Uninstall is a dummy implementation for non-Windows platforms.
func (s *Server) Uninstall(password string) error {
	return fmt.Errorf("Uninstall is only supported on Windows")
}
