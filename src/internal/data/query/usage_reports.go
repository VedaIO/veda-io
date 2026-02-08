package query

import (
	"database/sql"
	"time"
)

// AppUsageItem represents usage statistics for an application.
type AppUsageItem struct {
	ProcessName    string
	ExecutablePath string
	Count          int
}

// WebUsageItem represents usage statistics for a domain.
type WebUsageItem struct {
	Domain string
	Count  int
}

// ScreenTimeRecord represents duration spent in an application.
type ScreenTimeRecord struct {
	ExecutablePath  string
	DurationSeconds int
}

// GetAppUsageRanking returns the most used applications in a time range.
func GetAppUsageRanking(db *sql.DB, sinceTime, untilTime time.Time) ([]AppUsageItem, error) {
	q := "SELECT process_name, COUNT(*) as count FROM app_events WHERE 1=1"
	args := []interface{}{}

	if !sinceTime.IsZero() {
		q += " AND start_time >= ?"
		args = append(args, sinceTime.Unix())
	}
	if !untilTime.IsZero() {
		q += " AND start_time <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " GROUP BY process_name ORDER BY count DESC LIMIT 10"

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []AppUsageItem
	for rows.Next() {
		var item AppUsageItem
		if err := rows.Scan(&item.ProcessName, &item.Count); err != nil {
			continue
		}

		// Get the most recent exe path for this process name
		_ = db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1",
			item.ProcessName).Scan(&item.ExecutablePath)

		results = append(results, item)
	}
	return results, nil
}

// GetWebUsageRanking returns the most visited domains in a time range.
func GetWebUsageRanking(db *sql.DB, sinceTime, untilTime time.Time) ([]WebUsageItem, error) {
	q := `
		SELECT
			CASE
				WHEN INSTR(SUBSTR(url, INSTR(url, '//') + 2), '/') > 0
				THEN SUBSTR(url, INSTR(url, '//') + 2, INSTR(SUBSTR(url, INSTR(url, '//') + 2), '/') - 1)
				ELSE SUBSTR(url, INSTR(url, '//') + 2)
			END as domain,
			COUNT(*) as count
		FROM web_events
		WHERE 1=1
	`
	args := []interface{}{}

	if !sinceTime.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, sinceTime.Unix())
	}
	if !untilTime.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " GROUP BY domain ORDER BY count DESC LIMIT 10"

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []WebUsageItem
	for rows.Next() {
		var item WebUsageItem
		if err := rows.Scan(&item.Domain, &item.Count); err != nil {
			continue
		}
		results = append(results, item)
	}
	return results, nil
}

// GetScreenTimeTotals retrieves today's screen time grouped by application.
func GetScreenTimeTotals(db *sql.DB, todayStart int64) ([]ScreenTimeRecord, error) {
	q := `
		SELECT executable_path, SUM(duration_seconds) as total_duration
		FROM screen_time
		WHERE timestamp >= ?
		GROUP BY executable_path
		ORDER BY total_duration DESC
		LIMIT 10
	`
	rows, err := db.Query(q, todayStart)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []ScreenTimeRecord
	for rows.Next() {
		var item ScreenTimeRecord
		if err := rows.Scan(&item.ExecutablePath, &item.DurationSeconds); err != nil {
			continue
		}
		results = append(results, item)
	}
	return results, nil
}

// GetTotalDayScreenTime returns the absolute total screen time for today.
func GetTotalDayScreenTime(db *sql.DB, todayStart int64) (int, error) {
	var total int
	err := db.QueryRow(`
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM screen_time
		WHERE timestamp >= ?
	`, todayStart).Scan(&total)
	return total, err
}
