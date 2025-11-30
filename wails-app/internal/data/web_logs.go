package data

import (
	"database/sql"
	"time"
)

// GetWebLogs retrieves web logs from the database within a given time range.
// It returns a slice of string slices, where each inner slice represents a row with the following format:
// [Timestamp, URL]
func GetWebLogs(db *sql.DB, query, since, until string) ([][]string, error) {
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = ParseTime(since)
		if err != nil {
			return nil, err
		}
	}

	if until != "" {
		untilTime, err = ParseTime(until)
		if err != nil {
			return nil, err
		}
	}

	// Build the SQL query dynamically based on the provided time filters.
	q := "SELECT url, timestamp FROM web_events WHERE 1=1"
	args := make([]interface{}, 0)

	if query != "" {
		q += " AND url LIKE ?"
		args = append(args, "%"+query+"%")
	}

	if !sinceTime.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, sinceTime.Unix())
	}

	if !untilTime.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " ORDER BY timestamp DESC"

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			GetLogger().Printf("Failed to close rows: %v", err)
		}
	}()

	var entries [][]string
	for rows.Next() {
		var url string
		var timestamp int64
		if err := rows.Scan(&url, &timestamp); err != nil {
			continue
		}
		timestampStr := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		entries = append(entries, []string{timestampStr, url})
	}

	return entries, nil
}

// LogWebActivity records a visited URL in the database.
func LogWebActivity(url, title string, visitTime int64) error {
	db := GetDB()
	if db == nil {
		return nil // Or error if strict
	}

	// If visitTime is 0, use current time
	if visitTime == 0 {
		visitTime = time.Now().Unix()
	}

	query := "INSERT INTO web_events (url, timestamp) VALUES (?, ?)"
	_, err := db.Exec(query, url, visitTime)
	return err
}
