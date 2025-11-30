package api

import (
	"strings"
	"wails-app/internal/data"
)

// Search handles searches for application events.
func (s *Server) Search(query, since, until string) ([][]string, error) {
	return data.SearchAppEvents(s.db, strings.ToLower(query), since, until)
}
