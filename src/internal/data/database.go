package data

import (
	"database/sql"
	"fmt"
	"sync"

	"src/internal/data/connection"
	"src/internal/data/schema"
	"src/internal/data/write"
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
		globalDB, err = connection.OpenAndConfigureDB()
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
