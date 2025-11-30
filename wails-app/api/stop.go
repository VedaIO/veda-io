package api

import (
	"os"
	"time"
)

// Stop handles the graceful shutdown of the application.
func (s *Server) Stop() {
	s.Logger.Println("Received stop request. Shutting down...")

	go func() {
		time.Sleep(1 * time.Second)
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			s.Logger.Printf("Failed to close database: %v", err)
		}
		os.Exit(0)
	}()
}
