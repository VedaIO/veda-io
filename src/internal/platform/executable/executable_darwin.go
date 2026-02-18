//go:build darwin

package executable

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GetPublisherName returns the organization name from the code signature (Not Implemented).
func GetPublisherName(filePath string) (string, error) {
	return "", fmt.Errorf("not implemented on darwin")
}

// GetProductName returns the product name (Not Implemented).
func GetProductName(exePath string) (string, error) {
	return "", fmt.Errorf("not implemented on darwin")
}

// GetCommercialName retrieves the commercial name of the application.
// On macOS, this currently returns the filename without extension.
func GetCommercialName(exePath string) (string, error) {
	return strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath)), nil
}
