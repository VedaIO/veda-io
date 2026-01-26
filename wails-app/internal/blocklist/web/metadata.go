package web

import (
	"database/sql"
	"fmt"
	"wails-app/internal/data/logger"
	"wails-app/internal/data/query"
)

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
		meta, err := query.GetWebMetadata(db, domain)
		if err != nil {
			logger.GetLogger().Printf("Error querying web metadata for %s: %v", domain, err)
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
