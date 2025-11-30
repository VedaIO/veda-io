package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const appBlocklistFile = "blocklist.json"

// BlockedAppDetail represents the details of a blocked application.
type BlockedAppDetail struct {
	Name    string `json:"name"`
	ExePath string `json:"exe_path"`
}

// GetBlockedAppsWithDetails loads the blocklist and enriches it with the latest executable path from the database.
// This provides more context to the user in the UI.
func GetBlockedAppsWithDetails(db *sql.DB) ([]BlockedAppDetail, error) {
	names, err := LoadAppBlocklist()
	if err != nil {
		return nil, fmt.Errorf("could not load app blocklist names: %w", err)
	}

	if len(names) == 0 {
		return []BlockedAppDetail{}, nil
	}

	details := make([]BlockedAppDetail, 0, len(names))
	for _, name := range names {
		var exePath string
		// Find the most recent exe_path for the given process name to show the user the location of the blocked app.
		err := db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1", name).Scan(&exePath)
		if err != nil {
			if err == sql.ErrNoRows {
				// If no path is found in the database, we can still return the name.
				exePath = ""
			} else {
				// For other errors, log them but continue building the list.
				GetLogger().Printf("Error querying exe_path for %s: %v", name, err)
				exePath = ""
			}
		}
		details = append(details, BlockedAppDetail{Name: name, ExePath: exePath})
	}

	return details, nil
}

// LoadAppBlocklist reads the blocklist file from the user's cache directory.
// It returns a slice of strings, with all entries normalized to lowercase for case-insensitive matching.
// If the file doesn't exist, it returns an empty list, which is not considered an error.
func LoadAppBlocklist() ([]string, error) {
	cacheDir, _ := os.UserCacheDir()
	p := filepath.Join(cacheDir, "procguard", appBlocklistFile)

	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil // File not existing is not an error, just an empty list.
	}
	if err != nil {
		return nil, err
	}

	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocklist: %w", err)
	}

	// Normalize all entries to lowercase to ensure case-insensitive comparison.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// SaveAppBlocklist writes the given list of strings to the blocklist file.
// It normalizes all entries to lowercase before saving to ensure consistency.
// It also sets appropriate file permissions to secure the file.
func SaveAppBlocklist(list []string) error {
	// Normalize all entries to lowercase to ensure consistency.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, _ := os.UserCacheDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "procguard"), 0755)
	p := filepath.Join(cacheDir, "procguard", appBlocklistFile)

	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		return err
	}

	// Apply platform-specific file locking to prevent unauthorized modification.
	return platformLock(p) // build-tag dispatch
}

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

// ExportAppBlocklist saves the current blocklist to a user-specified file.
func ExportAppBlocklist(path string) error {
	list, err := LoadAppBlocklist()
	if err != nil {
		return err
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

// ImportAppBlocklist loads a blocklist from a file and merges it with the existing blocklist.
func ImportAppBlocklist(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}

	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}

	// The imported file can be a simple list of strings or a previously exported file.
	err = json.Unmarshal(content, &newEntries)
	if err != nil {
		err2 := json.Unmarshal(content, &savedList)
		if err2 != nil {
			return fmt.Errorf("load: invalid JSON format in %s", path)
		}
		newEntries = savedList.Blocked
	}

	existingList, err := LoadAppBlocklist()
	if err != nil {
		return err
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	return SaveAppBlocklist(existingList)
}
