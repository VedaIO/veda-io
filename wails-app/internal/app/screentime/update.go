package screentime

import (
	"wails-app/internal/data/logger"
	"wails-app/internal/data/write"
)

// flushScreenTime writes the buffered duration to the database.
// It updates the most recent record for the given app and title, adding the buffered duration.
func flushScreenTime(l logger.Logger, exePath, title string, duration int) {
	write.EnqueueWrite(`
		UPDATE screen_time 
		SET duration_seconds = duration_seconds + ?
		WHERE id = (
			SELECT id FROM screen_time 
			WHERE executable_path = ? AND window_title = ?
			ORDER BY timestamp DESC LIMIT 1
		)
	`, duration, exePath, title)
}
