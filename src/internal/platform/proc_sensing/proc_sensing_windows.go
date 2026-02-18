//go:build windows

package proc_sensing

/*
#cgo LDFLAGS: -lkernel32 -lpsapi
#include "proc_sensing_windows.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// GetAllProcesses returns a list of all active processes with its start and end time using the C WinAPI sensor.
func GetAllProcesses() ([]ProcessInfo, error) {
	cList := C.CaptureProcessSnapshot()
	defer C.FreeProcessSnapshot(cList)

	if cList.count == 0 || cList.processes == nil {
		return nil, nil
	}

	cProcs := unsafe.Slice(cList.processes, int(cList.count))
	results := make([]ProcessInfo, int(cList.count))

	for i, cp := range cProcs {
		results[i] = ProcessInfo{
			PID:           uint32(cp.pid),
			ParentPID:     uint32(cp.parent_pid),
			StartTimeNano: uint64(cp.start_time_nano),
			Name:          C.GoString(&cp.name[0]),
			ExePath:       C.GoString(&cp.exe_path[0]),
		}
	}

	return results, nil
}

// GetProcessByPID returns high-precision info for a specific PID efficiently.
func GetProcessByPID(pid uint32) (ProcessInfo, error) {
	cp := C.GetProcessInfoByPID(C.uint32_t(pid))
	if cp.start_time_nano == 0 {
		return ProcessInfo{}, fmt.Errorf("could not find or access process %d", pid)
	}

	return ProcessInfo{
		PID:           uint32(cp.pid),
		StartTimeNano: uint64(cp.start_time_nano),
		Name:          C.GoString(&cp.name[0]),
		ExePath:       C.GoString(&cp.exe_path[0]),
	}, nil
}
