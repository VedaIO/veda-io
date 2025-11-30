//go:build windows

package app

import (
	"fmt"
	"unsafe"
	"wails-app/internal/data"

	"golang.org/x/sys/windows"
)

const (
	// Integrity Level constants for Windows.
	// These are used to determine the trust level of a process.
	SECURITY_MANDATORY_UNTRUSTED_RID         = 0x00000000
	SECURITY_MANDATORY_LOW_RID               = 0x00001000
	SECURITY_MANDATORY_MEDIUM_RID            = 0x00002000
	SECURITY_MANDATORY_HIGH_RID              = 0x00003000
	SECURITY_MANDATORY_SYSTEM_RID            = 0x00004000
	SECURITY_MANDATORY_PROTECTED_PROCESS_RID = 0x00005000
)

// GetProcessIntegrityLevel returns the integrity level of a process on Windows.
// This is used to filter out high-privilege system processes that should not be monitored.
func GetProcessIntegrityLevel(pid uint32) (uint32, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		// Ignore errors for processes we can't open, as they are likely system processes
		// that we don't have permission to access anyway.
		return 0, nil
	}
	defer func() {
		if err := windows.Close(h); err != nil {
			data.GetLogger().Printf("Failed to close handle: %v", err)
		}
	}()

	var token windows.Token
	if err := windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token); err != nil {
		return 0, fmt.Errorf("could not open process token: %w", err)
	}
	defer func() {
		if err := token.Close(); err != nil {
			data.GetLogger().Printf("Failed to close token handle: %v", err)
		}
	}()

	// Get the required buffer size for the token information.
	var tokenInfoLen uint32
	_ = windows.GetTokenInformation(token, windows.TokenIntegrityLevel, nil, 0, &tokenInfoLen)
	if tokenInfoLen == 0 {
		return 0, fmt.Errorf("GetTokenInformation failed to get buffer size")
	}

	// Get the token information.
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
		// This can happen for certain SIDs, not necessarily an error, but no integrity level.
		return 0, nil
	}

	// The integrity level is the last sub-authority.
	integrityLevel := sid.SubAuthority(uint32(subAuthorityCount - 1))

	return integrityLevel, nil
}
