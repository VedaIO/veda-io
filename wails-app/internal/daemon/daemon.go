package daemon

import (
	"database/sql"
	"wails-app/internal/app"
	"wails-app/internal/app/screentime"
	"wails-app/internal/data/logger"
	"wails-app/internal/platform/autostart"
)

// Start initiates the background processes.
func Start(appLogger logger.Logger, db *sql.DB) {
	// Ensure the app starts on boot
	if _, err := autostart.EnsureAutostart(); err != nil {
		appLogger.Printf("Failed to set up autostart: %v", err)
	}
	// Start the process event logger to monitor process creation and termination.
	app.StartProcessEventLogger(appLogger, db)

	// Start the blocklist enforcer to kill blocked processes.
	app.StartBlocklistEnforcer(appLogger)

	// Start the screen time monitor to track foreground window usage.
	screentime.StartScreenTimeMonitor(appLogger, db)
}
