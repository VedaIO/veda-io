//go:build windows

package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"wails-app/internal/data"

	"golang.org/x/sys/windows/registry"
)

const (
	// HostName is the name of the native messaging host.
	// It must match the name specified in the browser extension's manifest.
	HostName = "com.nixuris.procguard"
)

// InstallNativeHost sets up the native messaging host for Chrome by creating a registry key
// that points to a manifest file. This allows the browser extension to communicate with the application.
func InstallNativeHost(exePath, extensionId string) error {
	log := data.GetLogger()
	keyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + HostName

	// Create the registry key that Chrome will use to find the native messaging host.
	k, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		log.Printf("Failed to create registry key: %v", err)
		return fmt.Errorf("failed to create registry key: %w", err)
	}
	defer func() {
		if err := k.Close(); err != nil {
			log.Printf("Failed to close registry key: %v", err)
		}
	}()

	// The manifest file must be stored in a location that the user has access to.
	// The user's cache directory is a good choice.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Failed to get user cache dir: %v", err)
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}
	appDataDir := filepath.Join(cacheDir, "procguard")
	configDir := filepath.Join(appDataDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("Failed to create config directory: %v", err)
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the manifest file that describes the native messaging host.
	manifestPath := filepath.Join(configDir, "native-host.json")
	if err := CreateManifest(manifestPath, exePath, extensionId); err != nil {
		log.Printf("Failed to create manifest file: %v", err)
		return fmt.Errorf("failed to create manifest file: %w", err)
	}

	// Set the default value of the registry key to the path of the manifest file.
	if err := k.SetStringValue("", manifestPath); err != nil {
		log.Printf("Failed to set registry key value: %v", err)
		return fmt.Errorf("failed to set registry key value: %w", err)
	}

	return nil
}

// CreateManifest creates the native messaging host manifest file.
// This file tells Chrome how to communicate with the native application.
func CreateManifest(manifestPath, exePath, extensionId string) error {
	manifest := map[string]interface{}{
		"name":        HostName,
		"description": "ProcGuard native messaging host",
		"path":        exePath,
		"type":        "stdio",
		"allowed_origins": []string{
			"chrome-extension://" + extensionId + "/",
		},
	}

	file, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			data.GetLogger().Printf("Failed to close file: %v", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

// RegisterExtension is a convenience function that gets the current executable's path
// and calls InstallNativeHost to register the extension.
func RegisterExtension(extensionId string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}
	return InstallNativeHost(exePath, extensionId)
}
