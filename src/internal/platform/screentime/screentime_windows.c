//go:build windows

#include "screentime_windows.h"
#include <windows.h>

int GetActiveWindowInfo(ActiveWindowInfo* info) {
    if (info == NULL) {
        return -1;
    }

    info->pid = 0;
    info->title[0] = L'\0';

    HWND hwnd = GetForegroundWindow();
    if (hwnd == NULL) {
        return -1;
    }

    DWORD pid = 0;
    GetWindowThreadProcessId(hwnd, &pid);
    if (pid == 0) {
        return -1;
    }
    info->pid = (uint32_t)pid;

    GetWindowTextW(hwnd, info->title, 256);

    return 0;
}
