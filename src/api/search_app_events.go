package api

import (
	"src/internal/data/query"
	"strings"
)

// Search handles searches for application events.
func (s *Server) Search(queryStr, since, until string) ([][]string, error) {
	return query.SearchAppEvents(s.db, strings.ToLower(queryStr), since, until)
}
