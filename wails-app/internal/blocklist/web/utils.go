package web

import (
	"fmt"
	"slices"
	"strings"
)

// AddWebsiteToBlocklist adds a domain to the web blocklist if it's not already there.
func AddWebsiteToBlocklist(domain string) (string, error) {
	list, err := LoadWebBlocklist()
	if err != nil {
		return "", err
	}

	lowerDomain := strings.ToLower(domain)
	if slices.Contains(list, lowerDomain) {
		return "exists", nil
	}

	list = append(list, lowerDomain)
	if err := SaveWebBlocklist(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "added", nil
}

// RemoveWebsiteFromBlocklist removes a domain from the web blocklist.
func RemoveWebsiteFromBlocklist(domain string) (string, error) {
	list, err := LoadWebBlocklist()
	if err != nil {
		return "", err
	}

	lowerDomain := strings.ToLower(domain)
	idx := slices.Index(list, lowerDomain)
	if idx == -1 {
		return "not found", nil
	}

	list = slices.Delete(list, idx, idx+1)
	if err := SaveWebBlocklist(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "removed", nil
}

// ClearWebBlocklist removes all entries from the web blocklist.
func ClearWebBlocklist() error {
	return SaveWebBlocklist([]string{})
}
