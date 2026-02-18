//go:build darwin

// Package integrity provides process integrity level functionality.
package process_integrity

// Integrity Level constants (Windows values, for cross-platform reference)
const (
	UntrustedRID        = 0x00000000
	LowRID              = 0x00001000
	MediumRID           = 0x00002000
	HighRID             = 0x00003000
	SystemRID           = 0x00004000
	ProtectedProcessRID = 0x00005000
)

// GetProcessLevel returns the integrity level of a process.
// macOS doesn't have Windows-style integrity levels, returns MediumRID.
func GetProcessLevel(pid uint32) (uint32, error) {
	return MediumRID, nil
}
