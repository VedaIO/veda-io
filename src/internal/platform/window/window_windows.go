//go:build windows

// Package window provides functionality to check window visibility.
package window

import (
	"syscall"
	"unsafe"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")

	enumWindowsCallback = syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		//nolint:govet
		params := (*enumWindowsParams)(unsafe.Pointer(lParam))
		var windowPid uint32
		_, _, err := procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&windowPid)))
		if err != syscall.Errno(0) {
			return 1 // Continue on error
		}

		if windowPid == params.pid {
			if isVisible, _, _ := procIsWindowVisible.Call(uintptr(hwnd)); isVisible != 0 {
				params.found = true
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue
	})
)

type enumWindowsParams struct {
	pid   uint32
	found bool
}

// HasVisibleWindow checks if a process with the given PID has a visible window.
func HasVisibleWindow(pid uint32) bool {
	params := &enumWindowsParams{pid: pid, found: false}
	_, _, _ = procEnumWindows.Call(enumWindowsCallback, uintptr(unsafe.Pointer(params)))
	return params.found
}
