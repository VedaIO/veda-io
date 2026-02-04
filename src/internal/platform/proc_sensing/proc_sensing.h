//go:build windows

#ifndef PROC_SENSING_H
#define PROC_SENSING_H

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

#ifdef __cplusplus
extern "C" {
#endif

// CaptureProcessSnapshot gathers all active processes via WinAPI.
ProcGuard_ProcessList CaptureProcessSnapshot();

// FreeProcessSnapshot releases the memory allocated by CaptureProcessSnapshot.
void FreeProcessSnapshot(ProcGuard_ProcessList list);

// GetProcessInfoByPID fetches high-precision info for a single PID without a full snapshot.
ProcGuard_ProcessInfo GetProcessInfoByPID(uint32_t pid);

#ifdef __cplusplus
}
#endif

#endif // PROC_SENSING_H
