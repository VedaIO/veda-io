package history

import (
	"wails-app/internal/data/write"
)

// This file contains functions for clearing history data from the database.
//
// DESIGN DECISION:
// We use write.EnqueueWrite instead of direct db.Exec to maintain data integrity and avoid
// race conditions with the background daemon which is constantly writing to these tables.
// write.EnqueueWrite uses a single-writer pattern via a Go channel, ensuring that delete operations
// are sequenced correctly with incoming log data.

// ClearAppHistory deletes all records related to application usage.
// This affects the following tables:
// 1. app_events: Contains process start/stop logs.
// 2. screen_time: Contains foreground window activity logs.
//
// After running this, all application-related leaderboards and history logs will be empty.
func ClearAppHistory() {
	// Delete all process event logs
	write.EnqueueWrite("DELETE FROM app_events")

	// Delete all screen time data
	write.EnqueueWrite("DELETE FROM screen_time")

	// VACUUM is not called here because it is a blocking operation and might be slow.
	// SQLite will reuse the free space for future inserts.
}

// ClearWebHistory deletes all records related to web browsing activity.
// This affects the following tables:
// 1. web_events: Contains the actual URL visit logs.
// 2. web_metadata: Contains cached titles and favicons for domains.
//
// After running this, all web-related leaderboards and history logs will be empty.
func ClearWebHistory() {
	// Delete all website visit logs
	write.EnqueueWrite("DELETE FROM web_events")

	// Delete all cached website metadata (titles, icons)
	write.EnqueueWrite("DELETE FROM web_metadata")
}
