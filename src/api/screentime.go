package api

import (
	"time"
)

// ScreenTimeItem represents a single application's screen time
type ScreenTimeItem struct {
	Name            string `json:"name"`            // Display name (commercial name or process name)
	ExecutablePath  string `json:"executablePath"`  // Full path to executable
	Icon            string `json:"icon"`            // Base64 encoded icon
	DurationSeconds int    `json:"durationSeconds"` // Total foreground time in seconds
}

// GetScreenTime retrieves today's screen time grouped by application
func (s *Server) GetScreenTime() ([]ScreenTimeItem, error) {
	// Get today's start timestamp (midnight)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	query := `
		SELECT executable_path, SUM(duration_seconds) as total_duration
		FROM screen_time
		WHERE timestamp >= ?
		GROUP BY executable_path
		ORDER BY total_duration DESC
		LIMIT 10
	`

	rows, err := s.db.Query(query, todayStart.Unix())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.Logger.Printf("Failed to close rows: %v", err)
		}
	}()

	var items []ScreenTimeItem
	for rows.Next() {
		var item ScreenTimeItem
		if err := rows.Scan(&item.ExecutablePath, &item.DurationSeconds); err != nil {
			continue
		}

		// Enrich with commercial name and icon
		commercialName, icon := s.getAppDetails(item.ExecutablePath)
		if commercialName != "" {
			item.Name = commercialName
		} else {
			// Extract filename from path as fallback
			item.Name = extractFileName(item.ExecutablePath)
		}
		item.Icon = icon

		items = append(items, item)
	}

	return items, nil
}

// extractFileName extracts the filename from a full path
func extractFileName(path string) string {
	// Handle both forward and backslashes
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// GetTotalScreenTime returns the total screen time for today in seconds
func (s *Server) GetTotalScreenTime() (int, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var total int
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(duration_seconds), 0)
		FROM screen_time
		WHERE timestamp >= ?
	`, todayStart.Unix()).Scan(&total)

	return total, err
}
