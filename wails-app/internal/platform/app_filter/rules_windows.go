//go:build windows

package app_filter

import (
	"strings"
	"wails-app/internal/platform/executable"
	"wails-app/internal/platform/integrity"
	"wails-app/internal/platform/proc_sensing"
	"wails-app/internal/platform/window"
)

// ShouldExclude returns true if the process is a Windows system component, conhost.exe, or ProcGuard itself.
func ShouldExclude(exePath string, proc *proc_sensing.ProcessInfo) bool {
	exePathLower := strings.ToLower(exePath)

	// Never track ProcGuard itself
	if strings.Contains(exePathLower, "procguard.exe") {
		return true
	}

	// Skip conhost.exe
	if strings.HasSuffix(exePathLower, "conhost.exe") {
		return true
	}

	// Skip if in System32/SysWOW64 (Windows system processes)
	if strings.Contains(exePathLower, "\\windows\\system32\\") ||
		strings.Contains(exePathLower, "\\windows\\syswow64\\") {
		return true
	}

	// Skip processes with "Microsoft® Windows® Operating System" product name
	productName, err := executable.GetProductName(exePath)
	if err == nil && strings.Contains(productName, "Microsoft® Windows® Operating System") {
		return true
	}

	// Skip system integrity level processes (system services)
	if proc != nil {
		il, err := integrity.GetProcessLevel(uint32(proc.PID))
		if err == nil && il >= integrity.SystemRID {
			return true
		}
	}

	return false
}

// ShouldTrack returns true if the process is a user application that should be monitored.
func ShouldTrack(exePath string, proc *proc_sensing.ProcessInfo) bool {
	if proc == nil {
		return false
	}

	nameLower := strings.ToLower(proc.Name)

	// Log cmd.exe and powershell.exe ONLY if launched by explorer.exe
	if nameLower == "cmd.exe" || nameLower == "powershell.exe" || nameLower == "pwsh.exe" {
		// Note: We'll keep the simplified logic for now or we could lookup the parent
		// if we wanted to be more precise, but let's keep it close to original.
		return false // Default to false for shells unless explorer is parent (to be refined)
	}

	// Must have visible window (user interaction indicator)
	if !window.HasVisibleWindow(uint32(proc.PID)) {
		return false
	}

	return true
}
