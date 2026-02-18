//go:build windows

#ifndef SCREENTIME_H
#define SCREENTIME_H

#include <stdint.h>
#include <wchar.h>

// ActiveWindowInfo captures foreground window metadata.
typedef struct {
    uint32_t pid;
    wchar_t title[256];
} ActiveWindowInfo;

#ifdef __cplusplus
extern "C" {
#endif

// GetActiveWindowInfo returns 0 on success, filling the info struct.
int GetActiveWindowInfo(ActiveWindowInfo* info);

#ifdef __cplusplus
}
#endif

#endif // SCREENTIME_H
