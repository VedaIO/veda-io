package app

import (
	"database/sql"
	"fmt"
	"wails-app/internal/data/logger"
)

// BlockedAppDetail represents the details of a blocked application.
// Move here from legacy app_blocklist.go
type BlockedAppDetail struct {
	Name    string `json:"name"`
	ExePath string `json:"exe_path"`
}

// GetBlockedAppsWithDetails loads the blocklist and enriches it with the latest executable path from the database.
// This provides more context to the user in the UI.
func GetBlockedAppsWithDetails(db *sql.DB) ([]BlockedAppDetail, error) {
	names, err := LoadAppBlocklist()
	if err != nil {
		return nil, fmt.Errorf("could not load app blocklist names: %w", err)
	}

	if len(names) == 0 {
		return []BlockedAppDetail{}, nil
	}

	details := make([]BlockedAppDetail, 0, len(names))
	for _, name := range names {
		var exePath string
		// Find the most recent exe_path for the given process name to show the user the location of the blocked app.
		err := db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1", name).Scan(&exePath)
		if err != nil {
			if err == sql.ErrNoRows {
				// If no path is found in the database, we can still return the name.
				exePath = ""
			} else {
				// For other errors, log them but continue building the list.
				logger.GetLogger().Printf("Error querying exe_path for %s: %v", name, err)
				exePath = ""
			}
		}
		details = append(details, BlockedAppDetail{Name: name, ExePath: exePath})
	}

	return details, nil
}
