//go:build windows

package proc_sensing

/*
#cgo LDFLAGS: -lkernel32 -lpsapi
#include <windows.h>
#include <tlhelp32.h>
#include <psapi.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

// ProcGuard_ProcessInfo matches our Go ProcessInfo struct for ABI compatibility.
typedef struct {
    uint32_t pid;
    uint32_t parent_pid;
    uint64_t start_time_nano;
    char name[260];
    char exe_path[260];
} ProcGuard_ProcessInfo;

// ProcGuard_ProcessList manages a collection of process snapshots.
typedef struct {
    ProcGuard_ProcessInfo* processes;
    uint32_t count;
} ProcGuard_ProcessList;

// CaptureProcessSnapshot implements the Windows-specific logic using WinAPI.
static inline ProcGuard_ProcessList CaptureProcessSnapshot() {
    ProcGuard_ProcessList list = { NULL, 0 };

    HANDLE snapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
    if (snapshot == INVALID_HANDLE_VALUE) return list;

    PROCESSENTRY32 entry;
    entry.dwSize = sizeof(PROCESSENTRY32);

    if (!Process32First(snapshot, &entry)) {
        CloseHandle(snapshot);
        return list;
    }

    // Count processes to allocate exact memory
    int count = 0;
    do { count++; } while (Process32Next(snapshot, &entry));

    list.processes = (ProcGuard_ProcessInfo*)calloc(count, sizeof(ProcGuard_ProcessInfo));
    if (!list.processes) {
        CloseHandle(snapshot);
        return list;
    }

    // Capture actual data
    if (Process32First(snapshot, &entry)) {
        int i = 0;
        do {
            ProcGuard_ProcessInfo* info = &list.processes[i];
            info->pid = entry.th32ProcessID;
            info->parent_pid = entry.th32ParentProcessID;

            memset(info->name, 0, sizeof(info->name));
            strncpy(info->name, entry.szExeFile, sizeof(info->name) - 1);

            // Fetch precise Start Time and Full Image Path
            HANDLE hProc = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, info->pid);
            if (hProc) {
                FILETIME ftCreate, ftExit, ftKernel, ftUser;
                if (GetProcessTimes(hProc, &ftCreate, &ftExit, &ftKernel, &ftUser)) {
                    unsigned long long rawTime = (((unsigned long long)ftCreate.dwHighDateTime) << 32) | ftCreate.dwLowDateTime;
                    info->start_time_nano = rawTime * 100; // 100ns steps to ns
                }

                DWORD pathSize = sizeof(info->exe_path);
                QueryFullProcessImageNameA(hProc, 0, info->exe_path, &pathSize);
                CloseHandle(hProc);
            }
            i++;
        } while (Process32Next(snapshot, &entry) && i < count);
        list.count = i;
    }

    CloseHandle(snapshot);
    return list;
}

static inline void FreeProcessSnapshot(ProcGuard_ProcessList list) {
    if (list.processes) free(list.processes);
}
*/
import "C"
import (
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
