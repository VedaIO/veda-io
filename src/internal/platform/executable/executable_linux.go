//go:build linux

package executable

import (
	"fmt"
	"path/filepath"
)

// GetPublisherName returns the organization name from the code signature (Not Implemented).
func GetPublisherName(filePath string) (string, error) {
	return "", fmt.Errorf("not implemented on linux")
}

// GetProductName returns the product name (Not Implemented).
func GetProductName(exePath string) (string, error) {
	return "", fmt.Errorf("not implemented on linux")
}

// GetCommercialName retrieves the commercial name of the application.
// On Linux, this currently returns the base filename.
func GetCommercialName(exePath string) (string, error) {
	return filepath.Base(exePath), nil
}
