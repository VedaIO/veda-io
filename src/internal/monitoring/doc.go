// Package monitoring provides a unified process monitoring system with subscriber-based architecture.
// It captures process snapshots at regular intervals and distributes them to registered subscribers.
//
// The package provides:
//   - MonitoringManager: Core component that orchestrates process monitoring with polling and recovery
//   - ProcessSubscriber: Interface for components that want to receive process snapshots
//   - Built-in subscribers: ProcessEventSubscriber, BlocklistSubscriber
//
// Usage:
//
//	manager := monitoring.NewMonitoringManager(logger, 2*time.Second)
//	manager.RegisterSubscriber(monitoring.NewProcessEventSubscriber(logger, repo))
//	manager.RegisterSubscriber(monitoring.NewBlocklistSubscriber(logger))
//	monitoring.SetGlobalManager(manager)
//	manager.Start()
//
//	// Later, to reset all subscribers:
//	monitoring.ResetGlobalManager()
package monitoring
