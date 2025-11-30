package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	globalDB *sql.DB
	dbOnce   sync.Once
	writeCh  chan WriteRequest
)

// WriteRequest represents a request to write to the database.
type WriteRequest struct {
	Query string
	Args  []interface{}
}

// InitDB initializes the database, creating the necessary tables and indexes if they don't exist.
// This function should be called once on application startup.
func InitDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		globalDB, err = openAndConfigureDB()
		if err != nil {
			return
		}

		if err = createSchema(globalDB); err != nil {
			err = fmt.Errorf("could not create schema: %w", err)
		}

		writeCh = make(chan WriteRequest, 100) // Buffered channel
		go StartDatabaseWriter(globalDB)
	})
	if err != nil {
		return nil, err
	}
	return globalDB, nil
}

// StartDatabaseWriter starts a goroutine that listens for write requests on the writeCh channel
// and executes them sequentially against the database.
func StartDatabaseWriter(db *sql.DB) {
	for req := range writeCh {
		_, err := db.Exec(req.Query, req.Args...)
		if err != nil {
			// If we can't write to the DB, log the failure.
			// We can't use the normal logger here as it might create a deadlock.
			log.Printf("[ERROR] Failed to execute write request: %v", err)
		}
	}
}

// EnqueueWrite sends a write request to the database writer channel.
func EnqueueWrite(query string, args ...interface{}) {
	writeCh <- WriteRequest{Query: query, Args: args}
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
	dbPath := filepath.Join(cacheDir, "procguard", "procguard.db")
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

// createSchema defines and executes the SQL statements to create the database schema.
func createSchema(db *sql.DB) error {
	schema := `
	-- app_events stores information about running processes.
	CREATE TABLE IF NOT EXISTS app_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		process_name TEXT NOT NULL,
		pid INTEGER NOT NULL,
		parent_process_name TEXT,
		exe_path TEXT,
		start_time INTEGER NOT NULL,
		end_time INTEGER
	);

	-- Indexes to speed up queries on app_events.
	CREATE INDEX IF NOT EXISTS idx_app_events_start_time ON app_events (start_time);
	CREATE INDEX IF NOT EXISTS idx_app_events_end_time ON app_events (end_time);
	CREATE INDEX IF NOT EXISTS idx_app_events_pid ON app_events (pid);

	-- web_events stores the URLs of visited websites.
	CREATE TABLE IF NOT EXISTS web_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	);

	-- Index to speed up queries on web_events.
	CREATE INDEX IF NOT EXISTS idx_web_events_timestamp ON web_events (timestamp);

	-- logs stores application logs.
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL
	);

	-- Index to speed up queries on logs.
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs (timestamp);

	-- web_metadata stores cached metadata for websites (title, icon).
	CREATE TABLE IF NOT EXISTS web_metadata (
		domain TEXT PRIMARY KEY,
		title TEXT,
		icon_url TEXT,
		timestamp INTEGER NOT NULL
	);
	`
	_, err := db.Exec(schema)
	return err
}
