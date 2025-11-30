//go:build windows

package api

import (
	"path/filepath"
	"strings"
	"wails-app/internal/app"

	"github.com/bi-zone/go-fileversion"
)

// getAppDetails retrieves details for a given application, such as its commercial name and icon.
func (srv *Server) getAppDetails(exePath string) (string, string) {
	// Get the commercial name from the executable's file version information.
	info, err := fileversion.New(exePath)
	var commercialName string
	if err == nil {
		commercialName = info.FileDescription()
		if commercialName == "" {
			commercialName = info.ProductName()
		}
		if commercialName == "" {
			commercialName = info.OriginalFilename()
		}
	}

	// If a commercial name could not be found, use the filename without the extension as a fallback.
	if commercialName == "" {
		commercialName = strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	}

	srv.iconCacheMu.Lock()
	icon, ok := srv.iconCache[exePath]
	srv.iconCacheMu.Unlock()
	if ok {
		return commercialName, icon
	}

	// Get the application's icon as a base64-encoded string.
	icon, err = app.GetAppIconAsBase64(exePath)
	if err != nil {
		// Log the error but don't fail the request, as the icon is not critical.
		srv.Logger.Printf("Failed to get icon for %s: %v", exePath, err)
	}

	srv.iconCacheMu.Lock()
	srv.iconCache[exePath] = icon
	srv.iconCacheMu.Unlock()

	return commercialName, icon
}
