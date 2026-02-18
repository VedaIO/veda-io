//go:build windows

#include "get_app_icon_windows.h"
#include <windows.h>
#include <shellapi.h>
#include <wingdi.h>
#include <stdlib.h>
#include <string.h>
#include <wchar.h>

ProcGuard_IconData* ExtractIconAsRGBA(const char* exePath) {
    // Convert UTF-8 path to wide string
    int wlen = MultiByteToWideChar(CP_UTF8, 0, exePath, -1, NULL, 0);
    if (wlen == 0) {
        return NULL;
    }
    
    wchar_t* wPath = (wchar_t*)malloc(wlen * sizeof(wchar_t));
    if (!wPath) {
        return NULL;
    }
    
    if (MultiByteToWideChar(CP_UTF8, 0, exePath, -1, wPath, wlen) == 0) {
        free(wPath);
        return NULL;
    }
    
    // Extract icon handle from executable
    HICON hIcon = ExtractIconW(NULL, wPath, 0);
    free(wPath);
    
    if (!hIcon) {
        return NULL;
    }

    // Get icon information
    ICONINFO iconInfo;
    if (!GetIconInfo(hIcon, &iconInfo)) {
        DestroyIcon(hIcon);
        return NULL;
    }

    // Get system icon size
    int width = GetSystemMetrics(SM_CXICON);
    int height = GetSystemMetrics(SM_CYICON);

    // Get device contexts
    HDC screenDC = GetDC(NULL);
    HDC memDC = CreateCompatibleDC(screenDC);

    // Setup BITMAPINFO for 32-bit RGBA
    BITMAPINFO bmi;
    memset(&bmi, 0, sizeof(bmi));
    bmi.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
    bmi.bmiHeader.biWidth = width;
    bmi.bmiHeader.biHeight = -height; // Top-down
    bmi.bmiHeader.biPlanes = 1;
    bmi.bmiHeader.biBitCount = 32;
    bmi.bmiHeader.biCompression = BI_RGB;

    // Allocate pixel buffer
    uint8_t* pixels = (uint8_t*)malloc(width * height * 4);
    if (!pixels) {
        DeleteDC(memDC);
        ReleaseDC(NULL, screenDC);
        DeleteObject(iconInfo.hbmColor);
        DeleteObject(iconInfo.hbmMask);
        DestroyIcon(hIcon);
        return NULL;
    }

    // Get the color bitmap bits
    if (!GetDIBits(memDC, iconInfo.hbmColor, 0, height, pixels, &bmi, DIB_RGB_COLORS)) {
        free(pixels);
        DeleteDC(memDC);
        ReleaseDC(NULL, screenDC);
        DeleteObject(iconInfo.hbmColor);
        DeleteObject(iconInfo.hbmMask);
        DestroyIcon(hIcon);
        return NULL;
    }

    // Convert BGRA to RGBA and apply alpha
    for (int i = 0; i < width * height; i++) {
        uint8_t b = pixels[i * 4 + 0];
        uint8_t g = pixels[i * 4 + 1];
        uint8_t r = pixels[i * 4 + 2];
        uint8_t a = pixels[i * 4 + 3];
        pixels[i * 4 + 0] = r;
        pixels[i * 4 + 1] = g;
        pixels[i * 4 + 2] = b;
        pixels[i * 4 + 3] = a;
    }

    // Cleanup GDI objects
    DeleteDC(memDC);
    ReleaseDC(NULL, screenDC);
    DeleteObject(iconInfo.hbmColor);
    DeleteObject(iconInfo.hbmMask);
    DestroyIcon(hIcon);

    // Allocate and return result
    ProcGuard_IconData* result = (ProcGuard_IconData*)malloc(sizeof(ProcGuard_IconData));
    if (!result) {
        free(pixels);
        return NULL;
    }

    result->width = (uint32_t)width;
    result->height = (uint32_t)height;
    result->pixels = pixels;

    return result;
}

void FreeIconData(ProcGuard_IconData* data) {
    if (data) {
        if (data->pixels) {
            free(data->pixels);
        }
        free(data);
    }
}
