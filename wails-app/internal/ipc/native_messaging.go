package ipc

import (
	"encoding/json"
	"net/http"
	"time"
	"wails-app/internal/data"
)

// HandleLogWebEvent handles requests from internal components to log a visited web URL.
func HandleLogWebEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.URL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	data.EnqueueWrite("INSERT INTO web_events (url, timestamp) VALUES (?, ?)", payload.URL, time.Now().Unix())
	w.WriteHeader(http.StatusOK)
}

// HandleLogWebMetadata handles requests from internal components to log web metadata.
func HandleLogWebMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload data.WebMetadata
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO web_metadata (domain, title, icon_url, timestamp)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(domain) DO UPDATE SET
			title = excluded.title,
			icon_url = excluded.icon_url,
			timestamp = excluded.timestamp;
	`
	data.EnqueueWrite(query, payload.Domain, payload.Title, payload.IconURL, time.Now().Unix())
	w.WriteHeader(http.StatusOK)
}

// HandleGetWebBlocklist handles requests from internal components to get the web blocklist.
func HandleGetWebBlocklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	list, err := data.LoadWebBlocklist()
	if err != nil {
		http.Error(w, "Failed to load web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		// Log the error, but don't return an HTTP error as the header might already be sent.
		data.GetLogger().Printf("Error encoding web blocklist response: %v", err)
	}
}
