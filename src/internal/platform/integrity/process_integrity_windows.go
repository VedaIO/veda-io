//go:build windows

// Package integrity provides process integrity level functionality.
package integrity

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Integrity Level constants for Windows.
const (
	UntrustedRID        = 0x00000000
	LowRID              = 0x00001000
	MediumRID           = 0x00002000
	HighRID             = 0x00003000
	SystemRID           = 0x00004000
	ProtectedProcessRID = 0x00005000
)

// GetProcessLevel returns the integrity level of a process on Windows.
// Returns 0 if the process cannot be opened (likely a system process).
func GetProcessLevel(pid uint32) (uint32, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		// Ignore errors for processes we can't open
		return 0, nil
	}
	defer func() { _ = windows.Close(h) }()

	var token windows.Token
	if err := windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token); err != nil {
		return 0, fmt.Errorf("could not open process token: %w", err)
	}
	defer func() { _ = token.Close() }()

	// Get the required buffer size
	var tokenInfoLen uint32
	_ = windows.GetTokenInformation(token, windows.TokenIntegrityLevel, nil, 0, &tokenInfoLen)
	if tokenInfoLen == 0 {
		return 0, fmt.Errorf("GetTokenInformation failed to get buffer size")
	}

	// Get the token information
	tokenInfo := make([]byte, tokenInfoLen)
	if err := windows.GetTokenInformation(token, windows.TokenIntegrityLevel, &tokenInfo[0], tokenInfoLen, &tokenInfoLen); err != nil {
		return 0, fmt.Errorf("could not get token information: %w", err)
	}

	til := (*windows.Tokenmandatorylabel)(unsafe.Pointer(&tokenInfo[0]))
	sid := til.Label.Sid

	if sid == nil {
		return 0, fmt.Errorf("SID is nil in token mandatory label")
	}

	subAuthorityCount := sid.SubAuthorityCount()
	if subAuthorityCount == 0 {
		return 0, nil
	}

	// The integrity level is the last sub-authority
	return sid.SubAuthority(uint32(subAuthorityCount - 1)), nil
}
