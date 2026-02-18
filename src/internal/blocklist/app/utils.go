package app

import (
	"fmt"
	"slices"
	"strings"
)

// AddAppToBlocklist adds a program to the blocklist if it's not already there.
func AddAppToBlocklist(name string) (string, error) {
	list, err := LoadAppBlocklist()
	if err != nil {
		return "", err
	}

	lowerName := strings.ToLower(name)
	if slices.Contains(list, lowerName) {
		return "exists", nil
	}

	list = append(list, lowerName)
	if err := SaveAppBlocklist(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "added", nil
}

// RemoveAppFromBlocklist removes a program from the blocklist.
func RemoveAppFromBlocklist(name string) (string, error) {
	list, err := LoadAppBlocklist()
	if err != nil {
		return "", err
	}

	lowerName := strings.ToLower(name)
	idx := slices.Index(list, lowerName)
	if idx == -1 {
		return "not found", nil
	}

	list = slices.Delete(list, idx, idx+1)
	if err := SaveAppBlocklist(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "removed", nil
}

// ClearAppBlocklist removes all entries from the blocklist.
func ClearAppBlocklist() error {
	return SaveAppBlocklist([]string{})
}
