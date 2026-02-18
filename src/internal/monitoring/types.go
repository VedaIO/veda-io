package monitoring

import (
	"src/internal/platform/proc_sensing"
	"time"
)

// ProcessSnapshot represents a point-in-time capture of all running processes.
// It is distributed to all registered subscribers on each polling interval.
type ProcessSnapshot struct {
	// Processes is the list of all running processes at the time of snapshot.
	Processes []proc_sensing.ProcessInfo
	// Timestamp is when this snapshot was captured.
	Timestamp time.Time
}

// ProcessSubscriber is the interface that subscribers must implement to receive process snapshots.
// Implementations should be stateless - all state should be maintained within the subscriber.
type ProcessSubscriber interface {
	// OnProcessesChanged is called with the current process snapshot on each polling interval.
	// The subscriber should process the snapshot and perform its actions.
	OnProcessesChanged(snapshot ProcessSnapshot)
	// Name returns the subscriber's name for logging purposes.
	Name() string
}

// ResettableSubscriber is an optional interface that subscribers can implement
// to support reset operations (e.g., clearing history).
type ResettableSubscriber interface {
	ProcessSubscriber
	// Reset clears the subscriber's internal state.
	Reset()
}
