package monitoring

import (
	"src/internal/data/logger"
	"src/internal/data/repository"
)

// StartDefault creates and starts a monitoring manager with all standard subscribers wired up.
// It is a convenience function for typical application startup.
func StartDefault(
	appLogger logger.Logger,
	appRepo *repository.AppRepository,
	screenTimeStarter func(logger.Logger, *repository.AppRepository, *repository.WebRepository),
) *MonitoringManager {
	manager := NewMonitoringManager(appLogger, DefaultPollingInterval)

	processEventSubscriber := NewProcessEventSubscriber(appLogger, appRepo)
	processEventSubscriber.InitializeFromDatabase()
	manager.RegisterSubscriber(processEventSubscriber)

	blocklistSubscriber := NewBlocklistSubscriber(appLogger)
	manager.RegisterSubscriber(blocklistSubscriber)

	SetGlobalManager(manager)
	manager.Start()

	if screenTimeStarter != nil {
		screenTimeStarter(appLogger, appRepo, nil)
	}

	return manager
}
