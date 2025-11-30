package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const webBlocklistFile = "web_blocklist.json"

// BlockedWebsiteDetail represents the details of a blocked website.
type BlockedWebsiteDetail struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"iconUrl"`
}

// GetBlockedWebsitesWithDetails loads the web blocklist and enriches it with metadata from the database.
func GetBlockedWebsitesWithDetails(db *sql.DB) ([]BlockedWebsiteDetail, error) {
	domains, err := LoadWebBlocklist()
	if err != nil {
		return nil, fmt.Errorf("could not load web blocklist domains: %w", err)
	}

	if len(domains) == 0 {
		return []BlockedWebsiteDetail{}, nil
	}

	details := make([]BlockedWebsiteDetail, 0, len(domains))
	for _, domain := range domains {
		meta, err := GetWebMetadata(db, domain)
		if err != nil {
			GetLogger().Printf("Error querying web metadata for %s: %v", domain, err)
			details = append(details, BlockedWebsiteDetail{Domain: domain})
			continue
		}
		if meta != nil {
			details = append(details, BlockedWebsiteDetail{
				Domain:  domain,
				Title:   meta.Title,
				IconURL: meta.IconURL,
			})
		} else {
			details = append(details, BlockedWebsiteDetail{Domain: domain})
		}
	}

	return details, nil
}

// LoadWebBlocklist reads the web blocklist file from the user's cache directory.
// It returns a slice of strings, with all entries normalized to lowercase for case-insensitive matching.
// If the file doesn't exist, it returns an empty list, which is not considered an error.
func LoadWebBlocklist() ([]string, error) {
	cacheDir, _ := os.UserCacheDir()
	p := filepath.Join(cacheDir, "procguard", webBlocklistFile)

	// If the blocklist file doesn't exist, return an empty list.
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal web blocklist: %w", err)
	}

	// Normalize all entries to lowercase for case-insensitive comparison.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// SaveWebBlocklist writes the given list of strings to the web blocklist file.
// It normalizes all entries to lowercase before saving to ensure consistency.
func SaveWebBlocklist(list []string) error {
	// Normalize all entries to lowercase to ensure consistency.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, _ := os.UserCacheDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "procguard"), 0755)
	p := filepath.Join(cacheDir, "procguard", webBlocklistFile)

	// Marshal the list to JSON with indentation for readability.
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal web blocklist: %w", err)
	}
	return os.WriteFile(p, b, 0600)
}

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
