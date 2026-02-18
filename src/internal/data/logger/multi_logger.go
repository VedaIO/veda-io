package logger

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"src/internal/data/write"
	"sync"
	"time"
)

// multiLogger is an implementation of the Logger interface that writes logs to multiple destinations:
// a log file and a SQLite database. This provides redundancy and flexible log analysis.
type multiLogger struct {
	db     *sql.DB
	file   *os.File
	logger *log.Logger
	mu     sync.Mutex
}

// Printf formats and logs a message with the INFO level.
func (l *multiLogger) Printf(format string, v ...interface{}) {
	l.write("INFO", fmt.Sprintf(format, v...))
}

// Fatalf formats and logs a message with the FATAL level, then exits the application.
func (l *multiLogger) Fatalf(format string, v ...interface{}) {
	l.write("FATAL", fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Println logs a message with the INFO level.
func (l *multiLogger) Println(v ...interface{}) {
	l.write("INFO", fmt.Sprintln(v...))
}

// Close safely closes the logger's resources (the log file and the database connection).
// This should be called before the application exits to ensure all logs are written.
func (l *multiLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			// If we can't close the file, there's not much we can do other than log it to stderr.
			log.Printf("Failed to close log file: %v", err)
		}
		l.file = nil
	}
	// Setting the db to nil prevents further writes to the database after it has been closed elsewhere.
	l.db = nil
}

// write is an internal method that writes a log message to all configured destinations.
func (l *multiLogger) write(level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write to the log file first.
	if l.logger != nil {
		l.logger.Printf("[%s] %s", level, message)
	}

	// Then, write to the database.
	if l.db != nil {
		write.EnqueueWrite("INSERT INTO logs (timestamp, level, message) VALUES (?, ?, ?)", time.Now().Unix(), level, message)
	}
}
