//go:build windows

// Package icon provides executable icon extraction functionality.
package icon

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

const (
	smCXIcon     = 11
	smCYIcon     = 12
	dibRGBColors = 0
	biRGB        = 0
)

// GetAppIconAsBase64 extracts the first icon from an executable
// and returns it as a base64-encoded PNG string.
func GetAppIconAsBase64(exePath string) (string, error) {
	largeIcons, err := extractIcons(exePath)
	if err != nil {
		return "", err
	}
	if len(largeIcons) == 0 {
		return "", fmt.Errorf("no icons found")
	}
	defer func() { _, _, _ = procDestroyIcon.Call(uintptr(largeIcons[0])) }()

	img, err := hIconToImage(largeIcons[0])
	if err != nil {
		return "", err
	}

	return imageToBase64(img)
}

func extractIcons(exePath string) ([]syscall.Handle, error) {
	pExePath, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	numIcons, _, _ := procExtractIconExW.Call(uintptr(unsafe.Pointer(pExePath)), uintptr(0xFFFFFFFF), 0, 0, 0)
	if numIcons == 0 {
		return nil, fmt.Errorf("no icons found in %s", exePath)
	}

	largeIcons := make([]syscall.Handle, numIcons)
	smallIcons := make([]syscall.Handle, numIcons)

	ret, _, err := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(pExePath)),
		0,
		uintptr(unsafe.Pointer(&largeIcons[0])),
		uintptr(unsafe.Pointer(&smallIcons[0])),
		uintptr(numIcons),
	)
	if ret == 0 {
		return nil, fmt.Errorf("ExtractIconExW failed: %w", err)
	}

	return largeIcons, nil
}

func getSystemMetrics(nIndex int) int32 {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(nIndex))
	return int32(ret)
}

func hIconToImage(hIcon syscall.Handle) (image.Image, error) {
	var iconInfo ICONINFO
	ret, _, err := procGetIconInfo.Call(uintptr(hIcon), uintptr(unsafe.Pointer(&iconInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("GetIconInfo failed: %w", err)
	}
	defer func() {
		_, _, _ = procDeleteObject.Call(uintptr(iconInfo.HbmColor))
		_, _, _ = procDeleteObject.Call(uintptr(iconInfo.HbmMask))
	}()

	width := getSystemMetrics(smCXIcon)
	height := getSystemMetrics(smCYIcon)

	screenDC, _, _ := procGetDC.Call(0)
	defer func() { _, _, _ = procReleaseDC.Call(0, screenDC) }()

	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	defer func() { _, _, _ = procDeleteDC.Call(memDC) }()

	var bmiColor BITMAPINFO
	bmiColor.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmiColor.BmiHeader))
	bmiColor.BmiHeader.BiWidth = width
	bmiColor.BmiHeader.BiHeight = -height
	bmiColor.BmiHeader.BiPlanes = 1
	bmiColor.BmiHeader.BiBitCount = 32
	bmiColor.BmiHeader.BiCompression = biRGB

	colorData := make([]byte, width*height*4)
	ret, _, err = procGetDIBits.Call(
		memDC,
		uintptr(iconInfo.HbmColor),
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&colorData[0])),
		uintptr(unsafe.Pointer(&bmiColor)),
		dibRGBColors,
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			offset := (y*int(width) + x) * 4
			img.SetRGBA(x, y, color.RGBA{
				R: colorData[offset+2],
				G: colorData[offset+1],
				B: colorData[offset],
				A: colorData[offset+3],
			})
		}
	}

	return img, nil
}

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
