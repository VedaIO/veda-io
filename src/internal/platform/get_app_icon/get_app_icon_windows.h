//go:build windows

#ifndef ICON_H
#define ICON_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    uint32_t width;
    uint32_t height;
    uint8_t* pixels; // RGBA format, width * height * 4 bytes
} ProcGuard_IconData;

ProcGuard_IconData* ExtractIconAsRGBA(const char* exePath);
void FreeIconData(ProcGuard_IconData* data);

#ifdef __cplusplus
}
#endif

#endif // ICON_H
