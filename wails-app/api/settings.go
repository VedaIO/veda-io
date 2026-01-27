package api

import (
	"fmt"
	"wails-app/internal/app"
	"wails-app/internal/app/screentime"
	"wails-app/internal/auth"
	"wails-app/internal/config"
	"wails-app/internal/data/history"
	"wails-app/internal/platform/autostart"
)

// GetAutostartStatus returns the current status of the autostart setting.
func (s *Server) GetAutostartStatus() (bool, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false, err
	}
	return cfg.AutostartEnabled, nil
}

// EnableAutostart enables the autostart feature for the application.
func (s *Server) EnableAutostart() error {
	_, err := autostart.EnsureAutostart()
	return err
}

// DisableAutostart disables the autostart feature for the application.
func (s *Server) DisableAutostart() error {
	return autostart.RemoveAutostart()
}

// ClearAppHistory removes all application usage logs and screen time data.
// This is irreversible and requires the user password.
func (s *Server) ClearAppHistory(password string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	history.ClearAppHistory()

	// Reset in-memory caches and states in the app logic
	// This ensures that previously logged applications can be re-logged
	// and screen time tracking starts fresh.
	app.ResetLoggedApps()
	screentime.ResetScreenTime()

	return nil
}

// ClearWebHistory removes all web browsing logs and cached website metadata.
// This is irreversible and requires the user password.
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
