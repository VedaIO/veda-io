package api

import (
	"fmt"
	"os"
	"strings"
	"wails-app/internal/auth"
	"wails-app/internal/blocklist/app"
	"wails-app/internal/config"
	"wails-app/internal/data/logger"
	"wails-app/internal/platform/autostart"
	"wails-app/internal/platform/nativehost"
	"wails-app/internal/platform/uninstall"

	"github.com/shirou/gopsutil/v3/process"
)

const appName = "ProcGuard"

// Uninstall handles the uninstallation of the application.
// It performs a series of cleanup tasks in a separate goroutine and then initiates a self-deletion process.
func (s *Server) Uninstall(password string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	go func() {
		// Close the logger and database to release file handles before deletion.
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			// We can't use the logger here, so just print to stderr.
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", err)
		}

		// Terminate any other running ProcGuard processes.
		killOtherProcGuardProcesses(s.Logger)

		// Unblock any files that were blocked by the application.
		if err := unblockAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unblock all files: %v\n", err)
		}

		// Perform other cleanup tasks.
		if err := autostart.RemoveAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove autostart: %v\n", err)
		}
		if err := nativehost.Remove(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove native host: %v\n", err)
		}

		// Initiate the self-deletion process using platform-specific logic.
		if err := uninstall.SelfDestruct(appName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initiate self-deletion: %v\n", err)
		}

		// Exit the application to allow the self-deletion to complete.
		os.Exit(0)
	}()

	return nil
}

// killOtherProcGuardProcesses finds and terminates any other running ProcGuard processes.
func killOtherProcGuardProcesses(logger logger.Logger) {
	currentPid := os.Getpid()
	procs, err := process.Processes()
	if err != nil {
		return
	}

	for _, p := range procs {
		if p.Pid == int32(currentPid) {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.HasPrefix(strings.ToLower(name), "procguard") {
			if err := p.Kill(); err != nil {
				logger.Printf("Failed to kill process %s: %v", name, err)
			}
		}
	}
}

// unblockAll restores the original names of any files that were blocked by the application.
func unblockAll() error {
	list, err := app.LoadAppBlocklist()
	if err != nil {
		return fmt.Errorf("could not load blocklist: %w", err)
	}

	for _, name := range list {
		if strings.HasSuffix(name, ".blocked") {
			newName := strings.TrimSuffix(name, ".blocked")
			if err := os.Rename(name, newName); err != nil {
				// Log the error but continue trying to unblock other files.
				logger.GetLogger().Printf("Failed to unblock file %s: %v", name, err)
			}
		}
	}

	return nil
}
