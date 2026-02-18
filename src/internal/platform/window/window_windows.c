//go:build windows

#include "window_windows.h"
#include <windows.h>

typedef struct {
    unsigned int target_pid;
    int found;
} SearchParams;

static BOOL CALLBACK enum_windows_callback(HWND hwnd, LPARAM lparam) {
    SearchParams* params = (SearchParams*)lparam;
    
    DWORD window_pid = 0;
    GetWindowThreadProcessId(hwnd, &window_pid);
    
    if (window_pid == params->target_pid) {
        if (IsWindowVisible(hwnd)) {
            params->found = 1;
            return FALSE;
        }
    }
    
    return TRUE;
}

int HasVisibleWindow(unsigned int pid) {
    SearchParams params = { pid, 0 };
    
    EnumWindows(enum_windows_callback, (LPARAM)&params);
    
    return params.found;
}
