package daemon

import (
	"database/sql"
	"wails-app/internal/app"
	"wails-app/internal/data"
)

// StartDaemon runs the core daemon logic as long-running background services.
func StartDaemon(appLogger data.Logger, db *sql.DB) {
	// Start the process event logger to monitor process creation and termination.
	app.StartProcessEventLogger(appLogger, db)

	// Start the blocklist enforcer to kill blocked processes.
	app.StartBlocklistEnforcer(appLogger)
}
