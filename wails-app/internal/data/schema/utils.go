package schema

import (
	"database/sql"
)

// CreateSchema defines and executes the SQL statements to create the database schema.
func CreateSchema(db *sql.DB) error {
	if _, err := db.Exec(AppSchema); err != nil {
		return err
	}

	// Migration: Ensure process_instance_key exists for existing databases.
	// We ignore the error because it will fail if the column already exists.
	_, _ = db.Exec("ALTER TABLE app_events ADD COLUMN process_instance_key TEXT")

	return nil
}
