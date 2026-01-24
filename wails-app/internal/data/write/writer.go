package write

import (
	"database/sql"
	"log"
)

// StartDatabaseWriter starts a goroutine that listens for write requests on the writeCh channel
// and executes them sequentially against the database.
func StartDatabaseWriter(db *sql.DB) {
	writeCh := GetWriteChannel()
	for req := range writeCh {
		_, err := db.Exec(req.Query, req.Args...)
		if err != nil {
			// If we can't write to the DB, log the failure.
			log.Printf("[ERROR] Failed to execute write request: %v", err)
		}
	}
}
