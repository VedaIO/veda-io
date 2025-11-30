package api

import (
	"database/sql"
	"sync"
	"wails-app/internal/data"
	"wails-app/internal/web"
)

// Server holds the dependencies for the API server, such as the database connection and the logger.
type Server struct {
	Logger          data.Logger
	IsAuthenticated bool
	Mu              sync.Mutex
	db              *sql.DB
	iconCache       map[string]string
	iconCacheMu     sync.Mutex
}

// NewServer creates a new Server with its dependencies.
func NewServer(db *sql.DB) *Server {
	return &Server{
		Logger:    data.GetLogger(),
		db:        db,
		iconCache: make(map[string]string),
	}
}

// AppDetailsResponse is the response for GetAppDetails.
type AppDetailsResponse struct {
	CommercialName string `json:"commercialName"`
	Icon           string `json:"icon"`
}

// GetAppDetails retrieves details for a given application, such as its commercial name and icon.
func (s *Server) GetAppDetails(exePath string) (AppDetailsResponse, error) {
	commercialName, icon := s.getAppDetails(exePath)

	response := AppDetailsResponse{
		CommercialName: commercialName,
		Icon:           icon,
	}
	return response, nil
}

// GetWebDetails retrieves metadata for a given domain.
func (s *Server) GetWebDetails(domain string) (data.WebMetadata, error) {
	meta, err := data.GetWebMetadata(s.db, domain)
	if err != nil {
		return data.WebMetadata{}, err
	}

	if meta == nil {
		// If no metadata is found, return an empty response.
		return data.WebMetadata{Domain: domain}, nil
	}

	return *meta, nil
}

// RegisterExtension handles the registration of the browser extension.
func (s *Server) RegisterExtension(id string) error {
	return web.RegisterExtension(id)
}
