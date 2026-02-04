package app

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"
)

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
