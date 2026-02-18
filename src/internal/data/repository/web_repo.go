package repository

import (
	"database/sql"
	"net/url"
	"src/internal/data/logger"
	"src/internal/data/write"
	"strings"
	"time"
)

// ExtractDomain extracts domain from a URL string.
// Returns the domain, or the original string if parsing fails.
func ExtractDomain(urlStr string) string {
	// Handle URLs without protocol
	if !strings.Contains(urlStr, "://") {
		urlStr = "http://" + urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	return u.Host
}

// WebBlockedDetail represents a domain with its recorded metadata.
type WebBlockedDetail struct {
	Domain  string
	Title   string
	IconURL string
}

// WebRepository handles all database operations related to web usage and metadata.
type WebRepository struct {
	db *sql.DB
}

// NewWebRepository creates a new instance of WebRepository.
func NewWebRepository(db *sql.DB) *WebRepository {
	return &WebRepository{db: db}
}

// GetUsageRanking returns the most visited domains in a time range.
func (r *WebRepository) GetUsageRanking(sinceTime, untilTime time.Time) ([]WebUsageItem, error) {
	q := `
		SELECT domain, COUNT(*) as count
		FROM web_events
		WHERE 1=1
	`
	args := []interface{}{}

	if !sinceTime.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, sinceTime.Unix())
	}
	if !untilTime.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " GROUP BY domain ORDER BY count DESC LIMIT 10"

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []WebUsageItem
	for rows.Next() {
		var item WebUsageItem
		if err := rows.Scan(&item.Domain, &item.Count); err != nil {
			continue
		}
		results = append(results, item)
	}
	return results, nil
}

// GetLogs retrieves web logs within a given time range.
func (r *WebRepository) GetLogs(queryStr, since, until string) ([][]string, error) {
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = ParseTime(since)
		if err != nil {
			return nil, err
		}
	}

	if until != "" {
		untilTime, err = ParseTime(until)
		if err != nil {
			return nil, err
		}
	}

	// Build the SQL query dynamically based on the provided time filters.
	// Returns: timestamp, domain, url
	q := "SELECT timestamp, domain, url FROM web_events WHERE 1=1"
	args := make([]interface{}, 0)

	if queryStr != "" {
		q += " AND domain LIKE ?"
		args = append(args, "%"+queryStr+"%")
	}

	if !sinceTime.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, sinceTime.Unix())
	}

	if !untilTime.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " ORDER BY timestamp DESC LIMIT 100"

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries [][]string
	for rows.Next() {
		var timestamp int64
		var domain string
		var url string
		if err := rows.Scan(&timestamp, &domain, &url); err != nil {
			continue
		}
		timestampStr := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		entries = append(entries, []string{timestampStr, domain, url})
	}

	return entries, nil
}

// GetBlockedDetails enriches a list of domains with metadata.
func (r *WebRepository) GetBlockedDetails(domains []string) ([]WebBlockedDetail, error) {
	details := make([]WebBlockedDetail, 0, len(domains))
	for _, domain := range domains {
		meta, err := r.GetMetadata(domain)
		if err != nil {
			logger.GetLogger().Printf("Repository Error: failed to query web meta for %s: %v", domain, err)
			details = append(details, WebBlockedDetail{Domain: domain})
			continue
		}

		if meta != nil {
			details = append(details, WebBlockedDetail{
				Domain:  domain,
				Title:   meta.Title,
				IconURL: meta.IconURL,
			})
		} else {
			details = append(details, WebBlockedDetail{Domain: domain})
		}
	}
	return details, nil
}

// GetMetadata retrieves cached metadata for a website.
func (r *WebRepository) GetMetadata(domain string) (*WebMetadata, error) {
	var meta WebMetadata
	err := r.db.QueryRow("SELECT domain, title, icon_url, timestamp FROM web_metadata WHERE domain = ?", domain).Scan(&meta.Domain, &meta.Title, &meta.IconURL, &meta.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

// SaveMetadata stores metadata for a domain.
func (r *WebRepository) SaveMetadata(domain, title, iconURL string) error {
	_, err := r.db.Exec(`
		INSERT INTO web_metadata (domain, title, icon_url, timestamp)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(domain) DO UPDATE SET
			title = excluded.title,
			icon_url = excluded.icon_url,
			timestamp = excluded.timestamp
	`, domain, title, iconURL, time.Now().Unix())
	return err
}

// LogWebEvent records a visit to a URL.
func (r *WebRepository) LogWebEvent(urlStr string) {
	domain := ExtractDomain(urlStr)
	write.EnqueueWrite("INSERT INTO web_events (url, domain, timestamp) VALUES (?, ?, ?)",
		urlStr, domain, time.Now().Unix())
}
