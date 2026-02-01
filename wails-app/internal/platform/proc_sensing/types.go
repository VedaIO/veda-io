package proc_sensing

import (
	"fmt"
)

// ProcessInfo represents a snapshot of a process at a specific point in time.
type ProcessInfo struct {
	PID           uint32
	ParentPID     uint32
	StartTimeNano uint64
	Name          string
	ExePath       string
}

// UniqueKey returns a string that uniquely identifies this specific process instance.
// This composite key solves the PID recycling issue by combining the PID with
// the high-precision start time.
func (p ProcessInfo) UniqueKey() string {
	return fmt.Sprintf("%d-%d", p.PID, p.StartTimeNano)
}
