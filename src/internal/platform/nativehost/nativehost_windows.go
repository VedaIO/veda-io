//go:build windows

package nativehost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
	"src/internal/data/logger"
)

const (
	// HostName is the name of the native messaging host.
	// It must match the name specified in the browser extension's manifest.
	HostName = "com.infraflakes.procguard"
)

// InstallNativeHost sets up the native messaging host for Chrome, Edge, and Firefox by creating registry keys
// that point to a manifest file. This allows browser extensions to communicate with the application.
func InstallNativeHost(exePath, extensionId string) error {
	log := logger.GetLogger()

	// Register for multiple browsers
	browsers := []string{
		`SOFTWARE\Google\Chrome\NativeMessagingHosts\` + HostName,
		`SOFTWARE\Microsoft\Edge\NativeMessagingHosts\` + HostName,
		`SOFTWARE\Mozilla\NativeMessagingHosts\` + HostName,
	}

	// The manifest file must be stored in a location that the user has access to.
	// The user's cache directory is a good choice.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Failed to get user cache dir: %v", err)
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}
	appDataDir := filepath.Join(cacheDir, "ProcGuard")
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

	// Register in all browser registry locations
	for _, keyPath := range browsers {
		// Create the registry key that the browser will use to find the native messaging host.
		k, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
		if err != nil {
			log.Printf("Failed to create registry key %s: %v", keyPath, err)
			continue // Don't fail if one browser isn't installed
		}

		// Set the default value of the registry key to the path of the manifest file.
		if err := k.SetStringValue("", manifestPath); err != nil {
			log.Printf("Failed to set registry key value for %s: %v", keyPath, err)
			_ = k.Close()
			continue
		}

		if err := k.Close(); err != nil {
			log.Printf("Failed to close registry key: %v", err)
		}
	}

	return nil
}

// CreateManifest creates the native messaging host manifest file.
// This file tells browsers how to communicate with the native application.
func CreateManifest(manifestPath, exePath, extensionId string) error {
	manifest := map[string]interface{}{
		"name":        HostName,
		"description": "ProcGuard native messaging host",
		"path":        exePath,
		"type":        "stdio",
		"allowed_origins": []string{
			"chrome-extension://hkanepohpflociaodcicmmfbdaohpceo/", // Chrome Web Store
			"chrome-extension://gpaafgcbiejjpfdgmjglehboafdicdjb/", // Dev Extension ID
		},
		"allowed_extensions": []string{
			"procguard@infraflakes.com",
		},
	}

	file, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close file: %v", err)
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

// Remove removes the native messaging host configuration from the system.
func Remove() error {
	// Delete the registry key for the native messaging host.
	keyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + HostName
	if err := registry.DeleteKey(registry.CURRENT_USER, keyPath); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Delete the manifest file.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	appDataDir := filepath.Join(cacheDir, "ProcGuard")
	manifestPath := filepath.Join(appDataDir, "config", "native-host.json")

	// Delete the heartbeat file too
	heartbeatPath := filepath.Join(appDataDir, "extension_heartbeat")
	if err := os.Remove(heartbeatPath); err != nil && !os.IsNotExist(err) {
		logger.GetLogger().Printf("Failed to remove heartbeat: %v", err)
	}

	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
