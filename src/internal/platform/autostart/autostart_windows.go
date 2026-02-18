//go:build windows

package autostart

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"src/internal/config"
	"src/internal/data/logger"

	"golang.org/x/sys/windows/registry"
)

const appName = "ProcGuard"

// EnsureAutostart sets up the application to run automatically when the user logs on.
// This is achieved by creating a registry entry in the standard `Software\\Microsoft\\Windows\\CurrentVersion\\Run` key.
// The executable is copied to a persistent AppData location to prevent issues if the original is moved or deleted.
func EnsureAutostart() (string, error) {
	// The path to the executable in the persistent location.
	destPath, err := copyExecutableToAppData()
	if err != nil {
		return "", fmt.Errorf("failed to set up persistent executable: %w", err)
	}

	// Open the registry key for user-specific autostart applications.
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return destPath, fmt.Errorf("failed to open Run registry key: %w", err)
	}
	defer func() {
		if err := key.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close registry key: %v", err)
		}
	}()

	// Check if the autostart entry already exists and is correct.
	// The entry should point to the persistent executable path with the --background flag.
	currentPath, _, err := key.GetStringValue(appName)
	if err == nil && currentPath == destPath+" --background" {
		return destPath, nil // Entry already exists and is correct.
	}

	// Set the registry value to point to the executable, including the --background flag.
	// This flag is critical for launching the app in a non-interactive service mode on startup,
	// ensuring all background functionality is active without opening a GUI.
	if err := key.SetStringValue(appName, destPath+" --background"); err != nil {
		return destPath, fmt.Errorf("failed to set startup registry key: %w", err)
	}

	// Update the config file to reflect the change in autostart status.
	// This ensures the application's internal state matches the system's autostart configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config to update autostart status:", err)
	} else {
		cfg.AutostartEnabled = true
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config to update autostart status:", err)
		}
	}

	return destPath, nil
}

// RemoveAutostart removes the registry entry that starts the application on user logon.
func RemoveAutostart() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil // Key doesn't exist, so there's nothing to do.
		}
		return err
	}
	defer func() {
		if err := key.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close registry key: %v", err)
		}
	}()

	// Delete the registry value. If it doesn't exist, we can ignore the error.
	if err := key.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Update the config file to reflect the change in autostart status.
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config to update autostart status:", err)
	} else {
		cfg.AutostartEnabled = false
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config to update autostart status:", err)
		}
	}

	return nil
}

// copyExecutableToAppData copies the current executable to a persistent location in the user's LOCALAPPDATA directory.
// This ensures that the application can be started by the system even if the original executable is moved or deleted.
func copyExecutableToAppData() (string, error) {
	sourcePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %w", err)
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	destDir := filepath.Join(localAppData, appName)
	destPath := filepath.Join(destDir, "ProcGuard.exe")

	// If the file already exists at the destination, there's no need to copy it again.
	if _, err := os.Stat(destPath); err == nil {
		return destPath, nil
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("error creating destination directory: %w", err)
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("error opening source executable: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close source file: %v", err)
		}
	}()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("error creating destination executable: %w", err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			logger.GetLogger().Printf("Failed to close destination file: %v", err)
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("error copying executable: %w", err)
	}

	return destPath, nil
}
