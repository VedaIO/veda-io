package api

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"
	"wails-app/internal/blocklist/web"
	"wails-app/internal/data/query"
)

// GetWebBlocklist returns the list of blocked websites with their details.
func (s *Server) GetWebBlocklist() ([]web.BlockedWebsiteDetail, error) {
	return web.GetBlockedWebsitesWithDetails(s.db)
}

// AddWebBlocklist adds a domain to the web blocklist.
func (s *Server) AddWebBlocklist(domain string) error {
	_, err := web.AddWebsiteToBlocklist(domain)
	return err
}

// RemoveWebBlocklist removes a domain from the web blocklist.
func (s *Server) RemoveWebBlocklist(domain string) error {
	_, err := web.RemoveWebsiteFromBlocklist(domain)
	return err
}

// ClearWebBlocklist removes all domains from the web blocklist.
func (s *Server) ClearWebBlocklist() error {
	return web.ClearWebBlocklist()
}

// SaveWebBlocklist saves the current web blocklist to a file for export.
func (s *Server) SaveWebBlocklist() ([]byte, error) {
	list, err := web.LoadWebBlocklist()
	if err != nil {
		return nil, err
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	return json.MarshalIndent(header, "", "  ")
}

// LoadWebBlocklist loads a web blocklist from an uploaded file and merges it with the existing blocklist.
func (s *Server) LoadWebBlocklist(content []byte) error {
	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}

	err := json.Unmarshal(content, &newEntries)
	if err != nil {
		err2 := json.Unmarshal(content, &savedList)
		if err2 != nil {
			return fmt.Errorf("invalid JSON format in uploaded file")
		}
		newEntries = savedList.Blocked
	}

	existingList, err := web.LoadWebBlocklist()
	if err != nil {
		return err
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	return web.SaveWebBlocklist(existingList)
}

// GetWebLogs retrieves web logs from the database within a given time range.
func (s *Server) GetWebLogs(queryStr, since, until string) ([][]string, error) {
	return query.GetWebLogs(s.db, queryStr, since, until)
}
