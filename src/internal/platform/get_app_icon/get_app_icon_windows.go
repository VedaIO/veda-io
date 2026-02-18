//go:build windows

package icon

/*
#cgo LDFLAGS: -lshell32 -lgdi32 -luser32
#include "get_app_icon_windows.h"
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"unsafe"
)

type IconData struct {
	Width  uint32
	Height uint32
	Pixels []byte
}

func GetAppIconAsBase64(exePath string) (string, error) {
	cPath := C.CString(exePath)
	defer C.free(unsafe.Pointer(cPath))

	cData := C.ExtractIconAsRGBA(cPath)
	if cData == nil {
		return "", fmt.Errorf("failed to extract icon from %s", exePath)
	}
	defer C.FreeIconData(cData)

	width := int(cData.width)
	height := int(cData.height)
	pixelCount := width * height * 4

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	cPixels := unsafe.Slice(cData.pixels, pixelCount)

	for i := 0; i < pixelCount; i++ {
		img.Pix[i] = uint8(cPixels[i])
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode icon as PNG: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
