//go:build windows

package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// SelfDestruct executes a batch script that deletes the application files after the main process has exited.
func SelfDestruct(appName string) error {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	appDataDir := filepath.Join(localAppData, appName)

	// Create a temporary batch file in the system's temp directory.
	tempDir := os.TempDir()
	batchFileName := fmt.Sprintf("delete_%s_%d.bat", appName, time.Now().UnixNano())
	batchFilePath := filepath.Join(tempDir, batchFileName)

	// The batch script waits for a moment to ensure the main process has exited,
	// then deletes the application's data directory and finally deletes itself.
	// We also attempt to delete the Roaming folder if it exists (legacy/WebView data).
	batchContent := fmt.Sprintf(`
@echo off
timeout /t 2 /nobreak > nul
rmdir /s /q "%s"
rmdir /s /q "%%APPDATA%%\%s"
rmdir /s /q "%%APPDATA%%\%s.exe"
del "%s"
`, appDataDir, appName, appName, batchFilePath)

	err := os.WriteFile(batchFilePath, []byte(batchContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write batch file: %w", err)
	}

	// Execute the batch file in a new, detached process so it can run independently of the main application.
	cmd := exec.Command("cmd.exe", "/C", batchFilePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start batch process: %w", err)
	}

	return nil
}
