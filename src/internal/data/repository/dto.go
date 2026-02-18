package repository

// AppUsageItem represents usage statistics for an application.
type AppUsageItem struct {
	ProcessName    string
	ExecutablePath string
	Count          int
}

// WebUsageItem represents usage statistics for a domain.
type WebUsageItem struct {
	Domain string
	Count  int
}

// ScreenTimeRecord represents duration spent in an application.
type ScreenTimeRecord struct {
	ExecutablePath  string
	DurationSeconds int
}

// WebMetadata holds the cached metadata for a website.
type WebMetadata struct {
	Domain    string `json:"domain"`
	Title     string `json:"title"`
	IconURL   string `json:"iconUrl"`
	Timestamp int64  `json:"timestamp"`
}
