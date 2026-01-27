package screentime

import (
	"strings"
	"wails-app/internal/platform/executable"
	"wails-app/internal/platform/integrity"

	"github.com/shirou/gopsutil/v3/process"
)

// ShouldTrackApp determines if an application should be tracked
// This is a shared filter used by both process logging and screen time tracking.
// It filters out system processes, Windows components, and other non-user apps.
func ShouldTrackApp(exePath string, proc *process.Process) bool {
	exePathLower := strings.ToLower(exePath)

	// Rule 0: Never track ProcGuard itself
	if strings.Contains(exePathLower, "procguard.exe") {
		return false
	}

	// Rule 1: Skip if in System32/SysWOW64 (Windows system processes)
	// This catches Task Manager, Settings, etc.
	if strings.Contains(exePathLower, "\\windows\\system32\\") ||
		strings.Contains(exePathLower, "\\windows\\syswow64\\") {
		return false
	}

	// Rule 2: Skip processes with "Microsoft® Windows® Operating System" product name
	// This catches ApplicationFrameHost, SystemSettings, and other UWP system components
	productName, err := executable.GetProductName(exePath)
	if err == nil && strings.Contains(productName, "Microsoft® Windows® Operating System") {
		return false
	}

	// Rule 3: Skip system integrity level processes (system services)
	if proc != nil {
		il, err := integrity.GetProcessLevel(uint32(proc.Pid))
		if err == nil && il >= integrity.SystemRID {
			return false
		}
	}

	return true
}

// ShouldTrackAppStrict is a stricter version that also filters Microsoft-signed apps
// Use this if you want to exclude all Microsoft apps (Office, Edge, etc.)
func ShouldTrackAppStrict(exePath string, proc *process.Process) bool {
	if !ShouldTrackApp(exePath, proc) {
		return false
	}

	// Additional: Skip all Microsoft-signed processes
	if executable.IsMicrosoftSigned(exePath) {
		return false
	}

	return true
}
