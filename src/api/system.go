package api

import (
	"fmt"
	"os"
	"src/internal/app/screentime"
	"src/internal/auth"
	app_blocklist "src/internal/blocklist/app"
	"src/internal/config"
	"src/internal/data/history"
	"src/internal/monitoring"
	"src/internal/platform/autostart"
	"src/internal/platform/nativehost"
	"src/internal/platform/proc_sensing"
	"src/internal/platform/uninstall"
	"src/internal/web/native_messaging"
	"strings"
	"time"
)

const appName = "ProcGuard"

// --- Lifecycle ---

func (s *Server) Shutdown() {
	s.Logger.Println("Received stop request. Shutting down...")
	native_messaging.Stop()

	go func() {
		time.Sleep(1 * time.Second)
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", err)
		}
		os.Exit(0)
	}()
}

func (s *Server) Uninstall(password string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	go func() {
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", err)
		}

		s.killOtherProcGuardProcesses()

		if err := s.unblockAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unblock all files: %v\n", err)
		}

		_ = autostart.RemoveAutostart()
		_ = nativehost.Remove()

		if err := uninstall.SelfDestruct(appName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initiate self-deletion: %v\n", err)
		}
		os.Exit(0)
	}()
	return nil
}

// --- Settings & History ---

func (s *Server) GetAutostartStatus() (bool, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false, err
	}
	return cfg.AutostartEnabled, nil
}

func (s *Server) EnableAutostart() error {
	_, err := autostart.EnsureAutostart()
	return err
}

func (s *Server) DisableAutostart() error {
	return autostart.RemoveAutostart()
}

func (s *Server) ClearAppHistory(password string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	history.ClearAppHistory()
	monitoring.ResetGlobalManager()
	screentime.ResetScreenTime()
	return nil
}

func (s *Server) ClearWebHistory(password string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	history.ClearWebHistory()
	return nil
}

// --- Internal Helpers ---

func (s *Server) killOtherProcGuardProcesses() {
	currentPid := os.Getpid()
	procs, err := proc_sensing.GetAllProcesses()
	if err != nil {
		return
	}

	for _, p := range procs {
		if int(p.PID) == currentPid {
			continue
		}
		if strings.HasPrefix(strings.ToLower(p.Name), "procguard") {
			if osProc, err := os.FindProcess(int(p.PID)); err == nil {
				_ = osProc.Kill()
			}
		}
	}
}

func (s *Server) unblockAll() error {
	list, err := app_blocklist.LoadAppBlocklist()
	if err != nil {
		return err
	}

	for _, name := range list {
		if strings.HasSuffix(name, ".blocked") {
			newName := strings.TrimSuffix(name, ".blocked")
			_ = os.Rename(name, newName)
		}
	}
	return nil
}
