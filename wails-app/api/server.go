package api

import (
	"database/sql"
	"sync"
	"wails-app/internal/data/logger"
	"wails-app/internal/data/query"
	"wails-app/internal/platform/executable"
	"wails-app/internal/platform/icon"
	"wails-app/internal/platform/nativehost"
)

// Server holds the dependencies for the API server, such as the database connection and the logger.
type Server struct {
	Logger          logger.Logger
	IsAuthenticated bool
	Mu              sync.Mutex
	db              *sql.DB
	iconCache       map[string]string
	iconCacheMu     sync.Mutex
}

// NewServer creates a new Server with its dependencies.
func NewServer(db *sql.DB) *Server {
	return &Server{
		Logger:    logger.GetLogger(),
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
func (s *Server) GetWebDetails(domain string) (query.WebMetadata, error) {
	meta, err := query.GetWebMetadata(s.db, domain)
	if err != nil {
		return query.WebMetadata{}, err
	}

	if meta == nil {
		// If no metadata is found, return an empty response.
		return query.WebMetadata{Domain: domain}, nil
	}

	return *meta, nil
}

// RegisterExtension handles the registration of the browser extension.
func (s *Server) RegisterExtension(id string) error {
	return nativehost.RegisterExtension(id)
}

// getAppDetails retrieves details for a given application, such as its commercial name and icon.
// It relies on platform-specific implementations in internal/platform.
func (s *Server) getAppDetails(exePath string) (string, string) {
	// Get the commercial name using platform-specific logic
	commercialName, err := executable.GetCommercialName(exePath)
	if err != nil {
		s.Logger.Printf("Failed to get commercial name for %s: %v", exePath, err)
		commercialName = "" // Fallback handled by UI or empty string
	}

	s.iconCacheMu.Lock()
	iconBase64, ok := s.iconCache[exePath]
	s.iconCacheMu.Unlock()
	if ok {
		return commercialName, iconBase64
	}

	// Get the application's icon as a base64-encoded string using platform-specific logic
	iconBase64, err = icon.GetAppIconAsBase64(exePath)
	if err != nil {
		// Log the error but don't fail the request, as the icon is not critical.
		s.Logger.Printf("Failed to get icon for %s: %v", exePath, err)
	}

	s.iconCacheMu.Lock()
	s.iconCache[exePath] = iconBase64
	s.iconCacheMu.Unlock()

	return commercialName, iconBase64
}
