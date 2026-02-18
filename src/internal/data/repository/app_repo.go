package repository

import (
	"database/sql"
	"fmt"
	"src/internal/data/logger"
	"src/internal/data/write"
	"strconv"
	"time"
)

// AppBlockedDetail represents an application name and its latest path.
type AppBlockedDetail struct {
	Name    string
	ExePath string
}

// AppRepository handles all database operations related to application usage and blocklists.
type AppRepository struct {
	db *sql.DB
}

// NewAppRepository creates a new instance of AppRepository.
func NewAppRepository(db *sql.DB) *AppRepository {
	return &AppRepository{db: db}
}

// GetUsageRanking returns the most used applications in a time range.
func (r *AppRepository) GetUsageRanking(sinceTime, untilTime time.Time) ([]AppUsageItem, error) {
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

	rows, err := r.db.Query(q, args...)
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
		_ = r.db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1",
			item.ProcessName).Scan(&item.ExecutablePath)

		results = append(results, item)
	}
	return results, nil
}

// GetScreenTimeTotals retrieves today's screen time grouped by application.
func (r *AppRepository) GetScreenTimeTotals(todayStart int64) ([]ScreenTimeRecord, error) {
	q := `
		SELECT executable_path, SUM(duration_seconds) as total_duration
		FROM screen_time
		WHERE timestamp >= ?
		GROUP BY executable_path
		ORDER BY total_duration DESC
		LIMIT 10
	`
	rows, err := r.db.Query(q, todayStart)
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
func (r *AppRepository) GetTotalDayScreenTime(todayStart int64) (int, error) {
	var total int
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM screen_time
		WHERE timestamp >= ?
	`, todayStart).Scan(&total)
	return total, err
}

// SearchEvents searches for application events in the database.
func (r *AppRepository) SearchEvents(queryStr, since, until string) ([][]string, error) {
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = ParseTime(since)
		if err != nil {
			return nil, fmt.Errorf("could not parse 'since' time: %w", err)
		}
	}

	if until != "" {
		untilTime, err = ParseTime(until)
		if err != nil {
			return nil, fmt.Errorf("could not parse 'until' time: %w", err)
		}
	}

	// Build the SQL query dynamically based on the provided filters.
	q := "SELECT process_name, pid, parent_process_name, exe_path, start_time, end_time FROM app_events WHERE 1=1"
	args := make([]interface{}, 0)

	if queryStr != "" {
		q += " AND (process_name LIKE ? OR parent_process_name LIKE ?)"
		likeQuery := "%" + queryStr + "%"
		args = append(args, likeQuery, likeQuery)
	}

	if !sinceTime.IsZero() {
		sinceUnix := sinceTime.Unix()
		q += " AND (end_time IS NULL OR end_time >= ?)"
		args = append(args, sinceUnix)
	}

	if !untilTime.IsZero() {
		untilUnix := untilTime.Unix()
		q += " AND start_time <= ?"
		args = append(args, untilUnix)
	}

	q += " ORDER BY start_time DESC LIMIT 100"

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results [][]string
	for rows.Next() {
		var processName, parentProcessName, exePath string
		var pid int32
		var startTime, endTime sql.NullInt64

		if err := rows.Scan(&processName, &pid, &parentProcessName, &exePath, &startTime, &endTime); err != nil {
			continue
		}

		startTimeStr := time.Unix(startTime.Int64, 0).Format("2006-01-02 15:04:05")

		results = append(results, []string{
			startTimeStr,
			processName,
			strconv.Itoa(int(pid)),
			parentProcessName,
			exePath,
		})
	}

	return results, nil
}

// GetBlockedDetails enriches a list of process names with their latest executable paths.
func (r *AppRepository) GetBlockedDetails(names []string) ([]AppBlockedDetail, error) {
	if len(names) == 0 {
		return []AppBlockedDetail{}, nil
	}

	details := make([]AppBlockedDetail, 0, len(names))
	for _, name := range names {
		var exePath string
		err := r.db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1", name).Scan(&exePath)
		if err != nil && err != sql.ErrNoRows {
			logger.GetLogger().Printf("Repository Error: failed to query exe_path for %s: %v", name, err)
		}
		details = append(details, AppBlockedDetail{Name: name, ExePath: exePath})
	}
	return details, nil
}

// SaveScreenTime logs a chunk of active screen time for an application.
func (r *AppRepository) SaveScreenTime(exePath string, seconds int) error {
	_, err := r.db.Exec("INSERT INTO screen_time (executable_path, duration_seconds, timestamp) VALUES (?, ?, ?)",
		exePath, seconds, time.Now().Unix())
	return err
}

// LogAppEvent records an application start event.
func (r *AppRepository) LogAppEvent(name string, pid uint32, parentName, exePath string, startTime int64, key string) {
	write.EnqueueWrite("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time, process_instance_key) VALUES (?, ?, ?, ?, ?, ?)",
		name, pid, parentName, exePath, startTime, key)
}

// CloseAppEvent records an application stop event.
func (r *AppRepository) CloseAppEvent(key string, endTime int64) {
	write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE process_instance_key = ? AND end_time IS NULL",
		endTime, key)
}

// ActiveSession represents a process that is currently logged as running.
type ActiveSession struct {
	Name string
	Key  string
}

// GetActiveSessions retrieves all applications that haven't been closed yet.
func (r *AppRepository) GetActiveSessions() ([]ActiveSession, error) {
	rows, err := r.db.Query("SELECT process_name, process_instance_key FROM app_events WHERE end_time IS NULL AND process_instance_key IS NOT NULL")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var sessions []ActiveSession
	for rows.Next() {
		var s ActiveSession
		if err := rows.Scan(&s.Name, &s.Key); err == nil {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}

// UpdateScreenTime increments the duration of the most recent screen time record for an application.
func (r *AppRepository) UpdateScreenTime(exePath string, duration int) {
	now := time.Now().Unix()
	write.EnqueueWrite(`
		UPDATE screen_time 
		SET duration_seconds = duration_seconds + ?, timestamp = ?
		WHERE id = (
			SELECT id FROM screen_time 
			WHERE executable_path = ? AND timestamp > ?
			ORDER BY timestamp DESC LIMIT 1
		)
	`, duration, now, exePath, now-300)
}

// EnsureActiveScreenTimeRecord ensures there is a recent screen time record to update.
func (r *AppRepository) EnsureActiveScreenTimeRecord(exePath string) {
	now := time.Now().Unix()
	write.EnqueueWrite(`
		INSERT INTO screen_time (executable_path, timestamp, duration_seconds)
		SELECT ?, ?, 1
		WHERE NOT EXISTS (
			SELECT 1 FROM screen_time 
			WHERE executable_path = ? AND timestamp > ?
			ORDER BY timestamp DESC LIMIT 1
		)
	`, exePath, now, exePath, now-300)
}
