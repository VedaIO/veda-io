package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"wails-app/internal/data/schema"
	"wails-app/internal/data/write"

	_ "modernc.org/sqlite"
)

var (
	globalDB *sql.DB
	dbOnce   sync.Once
)

// InitDB initializes the database, creating the necessary tables and indexes if they don't exist.
// This function should be called once on application startup.
func InitDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		globalDB, err = openAndConfigureDB()
		if err != nil {
			return
		}

		if err = schema.CreateSchema(globalDB); err != nil {
			err = fmt.Errorf("could not create schema: %w", err)
		}

		go write.StartDatabaseWriter(globalDB)
	})
	if err != nil {
		return nil, err
	}
	return globalDB, nil
}

// OpenDB opens a connection to the SQLite database.
// It does not attempt to create the schema and is intended for clients that only need to read data.
func OpenDB() (*sql.DB, error) {
	return InitDB()
}

// GetDB returns the global database instance.
func GetDB() *sql.DB {
	return globalDB
}

// openAndConfigureDB is a helper function that handles the common logic for opening and configuring the database connection.
func openAndConfigureDB() (*sql.DB, error) {
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
