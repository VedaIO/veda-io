//go:build windows

package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ICONINFO contains information about an icon, including its constituent bitmaps.
// This structure is used by the GetIconInfo function.
type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  syscall.Handle
	HbmColor syscall.Handle
}

// BITMAPINFOHEADER contains information about the dimensions and color format of a bitmap.
type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// RGBQUAD defines the colors in a color table.
type RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

// BITMAPINFO defines the dimensions and color information for a bitmap.
type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD
}

// Lazily load the required Windows DLLs and procedures for performance.
var (
	modShell32 = windows.NewLazySystemDLL("shell32.dll")
	modUser32  = windows.NewLazySystemDLL("user32.dll")
	modGdi32   = windows.NewLazySystemDLL("gdi32.dll")

	procExtractIconExW     = modShell32.NewProc("ExtractIconExW")
	procDestroyIcon        = modUser32.NewProc("DestroyIcon")
	procGetIconInfo        = modUser32.NewProc("GetIconInfo")
	procGetSystemMetrics   = modUser32.NewProc("GetSystemMetrics")
	procGetDC              = modUser32.NewProc("GetDC")
	procReleaseDC          = modUser32.NewProc("ReleaseDC")
	procCreateCompatibleDC = modGdi32.NewProc("CreateCompatibleDC")
	procGetDIBits          = modGdi32.NewProc("GetDIBits")
	procDeleteObject       = modGdi32.NewProc("DeleteObject")
	procDeleteDC           = modGdi32.NewProc("DeleteDC")
)

// Constants for Windows GDI and User32 APIs.
const (
	SM_CXICON      = 11 // System metric index for icon width.
	SM_CYICON      = 12 // System metric index for icon height.
	DIB_RGB_COLORS = 0  // Color table contains literal RGB values.
	BI_RGB         = 0  // Uncompressed format.
)

// ExtractIcons extracts a specified number of icons from an executable file.
func ExtractIcons(exePath string, nIcons uint) ([]syscall.Handle, []syscall.Handle, error) {
	pExePath, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	// First, call ExtractIconExW with nIcons = 0 to get the total number of icons in the file.
	numIcons, _, _ := procExtractIconExW.Call(uintptr(unsafe.Pointer(pExePath)), uintptr(0xFFFFFFFF), 0, 0, 0)
	if numIcons == 0 {
		return nil, nil, fmt.Errorf("no icons found in %s", exePath)
	}

	largeIcons := make([]syscall.Handle, numIcons)
	smallIcons := make([]syscall.Handle, numIcons)

	// Now, call ExtractIconExW again to actually extract the icons.
	ret, _, err := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(pExePath)),
		0,
		uintptr(unsafe.Pointer(&largeIcons[0])),
		uintptr(unsafe.Pointer(&smallIcons[0])),
		uintptr(numIcons),
	)
	if ret == 0 {
		return nil, nil, fmt.Errorf("ExtractIconExW failed: %w", err)
	}

	return largeIcons, smallIcons, nil
}

// GetSystemMetrics retrieves the specified system metric.
func GetSystemMetrics(nIndex int) int32 {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(nIndex))
	return int32(ret)
}

// HICONToImage converts an icon handle (HICON) to a Go `image.Image`.
// This is a complex process that involves several steps of Windows GDI manipulation.
func HICONToImage(hIcon syscall.Handle) (image.Image, error) {
	// 1. Get the icon information, which includes the color and mask bitmaps.
	var iconInfo ICONINFO
	ret, _, err := procGetIconInfo.Call(uintptr(hIcon), uintptr(unsafe.Pointer(&iconInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("GetIconInfo failed: %w", err)
	}
	// Ensure the bitmap handles are released when we're done.
	defer func() {
		_, _, _ = procDeleteObject.Call(uintptr(iconInfo.HbmColor))
		_, _, _ = procDeleteObject.Call(uintptr(iconInfo.HbmMask))
	}()

	// 2. Get the standard icon dimensions.
	width := GetSystemMetrics(SM_CXICON)
	height := GetSystemMetrics(SM_CYICON)

	// 3. Create a device context (DC) to work with.
	screenDC, _, _ := procGetDC.Call(0)
	defer func() { _, _, _ = procReleaseDC.Call(0, screenDC) }()

	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	defer func() { _, _, _ = procDeleteDC.Call(memDC) }()

	// 4. Set up the bitmap information header for the GetDIBits call.
	var bmiColor BITMAPINFO
	bmiColor.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmiColor.BmiHeader))
	bmiColor.BmiHeader.BiWidth = width
	bmiColor.BmiHeader.BiHeight = -height // A negative height indicates a top-down bitmap.
	bmiColor.BmiHeader.BiPlanes = 1
	bmiColor.BmiHeader.BiBitCount = 32 // We want a 32-bit RGBA bitmap.
	bmiColor.BmiHeader.BiCompression = BI_RGB

	// 5. Get the raw bitmap data.
	colorData := make([]byte, width*height*4)
	ret, _, err = procGetDIBits.Call(
		memDC,
		uintptr(iconInfo.HbmColor),
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&colorData[0])),
		uintptr(unsafe.Pointer(&bmiColor)),
		DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits (color data) failed: %w", err)
	}

	// 6. Create a Go `image.Image` and copy the bitmap data into it.
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			offset := (y*int(width) + x) * 4
			// The bitmap data is in BGRA order, so we need to swap the blue and red channels.
			img.SetRGBA(x, y, color.RGBA{R: colorData[offset+2], G: colorData[offset+1], B: colorData[offset], A: colorData[offset+3]})
		}
	}

	return img, nil
}

// imageToBase64 converts an `image.Image` to a base64-encoded PNG string.
func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// GetAppIconAsBase64 is a convenience function that extracts the first icon from an executable
// and returns it as a base64-encoded PNG string.
func GetAppIconAsBase64(exePath string) (string, error) {
	largeIcons, _, err := ExtractIcons(exePath, 1)
	if err != nil {
		return "", err
	}
	if len(largeIcons) == 0 {
		return "", fmt.Errorf("no icons found")
	}
	defer func() { _, _, _ = procDestroyIcon.Call(uintptr(largeIcons[0])) }()

	img, err := HICONToImage(largeIcons[0])
	if err != nil {
		return "", err
	}

	return imageToBase64(img)
}
