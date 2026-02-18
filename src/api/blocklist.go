package api

import (
	"encoding/json"
	"fmt"
	"slices"
	"src/internal/blocklist/app"
	"src/internal/blocklist/web"
	"strings"
	"time"
)

// --- App Blocklist ---

func (s *Server) BlockApps(names []string) error {
	list, err := app.LoadAppBlocklist()
	if err != nil {
		return err
	}
	for _, name := range names {
		lowerName := strings.ToLower(name)
		if !slices.Contains(list, lowerName) {
			list = append(list, lowerName)
		}
	}
	return app.SaveAppBlocklist(list)
}

func (s *Server) UnblockApps(names []string) error {
	list, err := app.LoadAppBlocklist()
	if err != nil {
		return err
	}
	for _, name := range names {
		lowerName := strings.ToLower(name)
		list = slices.DeleteFunc(list, func(item string) bool {
			return item == lowerName
		})
	}
	return app.SaveAppBlocklist(list)
}

func (s *Server) GetAppBlocklist() ([]app.BlockedAppDetail, error) {
	names, err := app.LoadAppBlocklist()
	if err != nil {
		return nil, err
	}

	records, err := s.Apps.GetBlockedDetails(names)
	if err != nil {
		return nil, err
	}

	details := make([]app.BlockedAppDetail, 0, len(records))
	for _, r := range records {
		details = append(details, app.BlockedAppDetail{
			Name:    r.Name,
			ExePath: r.ExePath,
		})
	}
	return details, nil
}

func (s *Server) ClearAppBlocklist() error {
	return app.ClearAppBlocklist()
}

func (s *Server) SaveAppBlocklist() ([]byte, error) {
	list, err := app.LoadAppBlocklist()
	if err != nil {
		return nil, err
	}
	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}
	return json.MarshalIndent(header, "", "  ")
}

func (s *Server) LoadAppBlocklist(content []byte) error {
	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}
	if err := json.Unmarshal(content, &newEntries); err != nil {
		if err2 := json.Unmarshal(content, &savedList); err2 != nil {
			return fmt.Errorf("invalid JSON format in uploaded file")
		}
		newEntries = savedList.Blocked
	}
	existingList, err := app.LoadAppBlocklist()
	if err != nil {
		return err
	}
	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}
	return app.SaveAppBlocklist(existingList)
}

// --- Web Blocklist ---

func (s *Server) GetWebBlocklist() ([]web.BlockedWebsiteDetail, error) {
	domains, err := web.LoadWebBlocklist()
	if err != nil {
		return nil, err
	}

	records, err := s.Web.GetBlockedDetails(domains)
	if err != nil {
		return nil, err
	}

	details := make([]web.BlockedWebsiteDetail, 0, len(records))
	for _, r := range records {
		details = append(details, web.BlockedWebsiteDetail{
			Domain:  r.Domain,
			Title:   r.Title,
			IconURL: r.IconURL,
		})
	}
	return details, nil
}

func (s *Server) AddWebBlocklist(domain string) error {
	_, err := web.AddWebsiteToBlocklist(domain)
	return err
}

func (s *Server) RemoveWebBlocklist(domain string) error {
	_, err := web.RemoveWebsiteFromBlocklist(domain)
	return err
}

func (s *Server) ClearWebBlocklist() error {
	return web.ClearWebBlocklist()
}

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

func (s *Server) LoadWebBlocklist(content []byte) error {
	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}
	if err := json.Unmarshal(content, &newEntries); err != nil {
		if err2 := json.Unmarshal(content, &savedList); err2 != nil {
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
