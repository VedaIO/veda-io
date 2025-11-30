package api

import (
	"wails-app/internal/daemon"
	"wails-app/internal/data"
)

// GetAutostartStatus returns the current status of the autostart setting.
func (s *Server) GetAutostartStatus() (bool, error) {
	cfg, err := data.LoadConfig()
	if err != nil {
		return false, err
	}
	return cfg.AutostartEnabled, nil
}

// EnableAutostart enables the autostart feature for the application.
func (s *Server) EnableAutostart() error {
	_, err := daemon.EnsureAutostart()
	return err
}

// DisableAutostart disables the autostart feature for the application.
func (s *Server) DisableAutostart() error {
	return daemon.RemoveAutostart()
}
