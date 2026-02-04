//go:build windows

#include "proc_sensing.h"
#include <windows.h>
#include <tlhelp32.h>
#include <psapi.h>
#include <stdlib.h>
#include <string.h>

// CaptureProcessSnapshot implements the Windows-specific logic using WinAPI.
ProcGuard_ProcessList CaptureProcessSnapshot() {
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

void FreeProcessSnapshot(ProcGuard_ProcessList list) {
    if (list.processes) free(list.processes);
}
// GetProcessInfoByPID fetches high-precision info for a single PID without a full snapshot. 
ProcGuard_ProcessInfo GetProcessInfoByPID(uint32_t pid) {
    ProcGuard_ProcessInfo info = { 0 };
    info.pid = pid;
    
    HANDLE hProc = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pid);
    if (!hProc) return info;

    FILETIME ftCreate, ftExit, ftKernel, ftUser;
    if (GetProcessTimes(hProc, &ftCreate, &ftExit, &ftKernel, &ftUser)) {
        unsigned long long rawTime = (((unsigned long long)ftCreate.dwHighDateTime) << 32) | ftCreate.dwLowDateTime;
        info.start_time_nano = rawTime * 100;
    }

    DWORD pathSize = sizeof(info.exe_path);
    if (QueryFullProcessImageNameA(hProc, 0, info.exe_path, &pathSize)) {
        // Extract basename for name field
        char* lastSlash = strrchr(info.exe_path, '\\');
        if (lastSlash) {
            strncpy(info.name, lastSlash + 1, sizeof(info.name) - 1);
        } else {
            strncpy(info.name, info.exe_path, sizeof(info.name) - 1);
        }
    }
    
    CloseHandle(hProc);
    return info;
}
