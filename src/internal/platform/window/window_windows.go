//go:build windows

package window

/*
#cgo LDFLAGS: -luser32
#include "window_windows.h"
*/
import "C"

func HasVisibleWindow(pid uint32) bool {
	return C.HasVisibleWindow(C.uint(pid)) != 0
}
