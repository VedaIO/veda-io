package logger

// Logger defines the interface for the application's logger.
type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Println(v ...interface{})
	Close()
}
