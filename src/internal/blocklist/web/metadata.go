package web

// BlockedWebsiteDetail represents the details of a blocked website.
type BlockedWebsiteDetail struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"iconUrl"`
}
