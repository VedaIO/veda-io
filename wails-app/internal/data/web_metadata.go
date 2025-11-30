package data

import (
	"database/sql"
)

// WebMetadata holds the cached metadata for a website, such as its title and icon URL.
type WebMetadata struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"iconUrl"`
}

// GetWebMetadata retrieves the cached metadata for a given domain from the database.
func GetWebMetadata(db *sql.DB, domain string) (*WebMetadata, error) {
	query := "SELECT title, icon_url FROM web_metadata WHERE domain = ?"
	row := db.QueryRow(query, domain)

	var meta WebMetadata
	meta.Domain = domain
	if err := row.Scan(&meta.Title, &meta.IconURL); err != nil {
		if err == sql.ErrNoRows {
			// It's not an error if no metadata is found for a domain; it simply means we haven't cached it yet.
			// In this case, we return nil for both the metadata and the error.
			return nil, nil
		}
		return nil, err
	}

	return &meta, nil
}
