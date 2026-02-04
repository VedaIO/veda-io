package screentime

import (
	"src/internal/data/logger"
	"src/internal/data/write"
	"time"
)

// flushScreenTime writes the buffered duration to the database.
// It updates the most recent record for the given app, adding the buffered duration.
func flushScreenTime(l logger.Logger, exePath string, duration int) {
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
