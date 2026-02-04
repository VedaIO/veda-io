package connection

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenAndConfigureDB handles the common logic for opening and configuring the database connection.
func OpenAndConfigureDB() (*sql.DB, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user cache dir: %w", err)
	}
	dbPath := filepath.Join(cacheDir, "ProcGuard", "procguard.db")
	log.Printf("Database path: %s", dbPath)

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Printf("Error creating database directory: %v", err)
		return nil, fmt.Errorf("could not create database directory: %w", err)
	}

	// Enable Write-Ahead Logging (WAL) mode. WAL allows for higher concurrency by separating read and write operations,
	// which is beneficial for this application where the daemon is constantly writing and the API server is reading.
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", dbPath))
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	return db, nil
}
