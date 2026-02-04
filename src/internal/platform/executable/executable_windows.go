//go:build windows

package executable

import (
	"debug/pe"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bi-zone/go-fileversion"
	"go.mozilla.org/pkcs7"
)

// GetPublisherName returns the organization name from the authenticode signature of the executable.
func GetPublisherName(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer func() { _ = file.Close() }()

	peFile, err := pe.NewFile(file)
	if err != nil {
		return "", fmt.Errorf("error parsing PE file %s: %w", filePath, err)
	}

	var securityDir pe.DataDirectory
	switch oh := peFile.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		securityDir = oh.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY]
	case *pe.OptionalHeader64:
		securityDir = oh.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY]
	default:
		return "", fmt.Errorf("unsupported PE optional header type: %T", peFile.OptionalHeader)
	}

	if securityDir.Size == 0 {
		return "", fmt.Errorf("no security directory found")
	}

	pkcs7Offset := int64(securityDir.VirtualAddress + 8)
	pkcs7Size := int64(securityDir.Size - 8)

	if pkcs7Size <= 0 {
		return "", fmt.Errorf("invalid signature size")
	}

	signatureBytes := make([]byte, pkcs7Size)
	_, err = file.ReadAt(signatureBytes, pkcs7Offset)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading signature from file: %w", err)
	}

	p7, err := pkcs7.Parse(signatureBytes)
	if err != nil {
		return "", fmt.Errorf("error parsing PKCS#7 signature: %w", err)
	}

	if len(p7.Certificates) == 0 {
		return "", fmt.Errorf("no certificates found in signature")
	}

	for _, cert := range p7.Certificates {
		if len(cert.Subject.Organization) > 0 {
			return cert.Subject.Organization[0], nil
		}
	}

	return "", fmt.Errorf("no organization name found in any certificate")
}

// GetProductName returns the product name from the version info resource.
func GetProductName(exePath string) (string, error) {
	info, err := fileversion.New(exePath)
	if err != nil {
		return "", err
	}
	return info.ProductName(), nil
}

// GetCommercialName retrieves the commercial name of the application.
// It tries to find the FileDescription, ProductName, or OriginalFilename from the executable's version info.
// If none are found, it falls back to the filename without extension.
func GetCommercialName(exePath string) (string, error) {
	info, err := fileversion.New(exePath)
	var commercialName string
	if err == nil {
		commercialName = info.FileDescription()
		if commercialName == "" {
			commercialName = info.ProductName()
		}
		if commercialName == "" {
			commercialName = info.OriginalFilename()
		}
	}

	// Fallback to filename
	if commercialName == "" {
		commercialName = strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	}

	return commercialName, nil
}
