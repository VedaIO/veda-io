//go:build !windows

package api

import (
	"path/filepath"
	"strings"
)

// getAppDetails is a dummy implementation for non-Windows platforms.
func (srv *Server) getAppDetails(exePath string) (string, string) {
	commercialName := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	return commercialName, ""
}
