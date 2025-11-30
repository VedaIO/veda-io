package web

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
	"wails-app/internal/data"
)

const (
	// pollInterval is the interval at which the web blocklist is polled for changes.
	pollInterval = 500 * time.Millisecond
	internalAPI  = "http://127.0.0.1:58142"
)

// WebMetadataPayload is the payload for the log_web_metadata message from the extension.
type WebMetadataPayload struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"iconUrl"`
}

// Request is a message received from the browser extension.
type Request struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Response is a message sent to the browser extension.
type Response struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Run starts the native messaging host, which listens for messages from the browser extension.
func Run() {
	log := data.GetLogger()

	// Start a goroutine to poll the web blocklist and push updates to the extension.
	go pollWebBlocklist()

	// The main loop reads messages from stdin, which is connected to the browser extension.
	for {
		// The native messaging protocol prefixes each message with its length in bytes.
		var length uint32
		if err := binary.Read(os.Stdin, binary.LittleEndian, &length); err != nil {
			if err == io.EOF {
				log.Println("EOF received, exiting native messaging host.")
				break // Exit loop on EOF
			}
			log.Printf("Error reading message length: %v", err)
			continue
		}

		msg := make([]byte, length)
		if _, err := io.ReadFull(os.Stdin, msg); err != nil {
			log.Printf("Error reading message body: %v", err)
			continue
		}

		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Handle the message based on its type.
		switch req.Type {
		case "ping":
			var payload string
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling ping payload: %v", err)
				continue
			}
			resp := Response{
				Type:    "echo",
				Payload: payload,
			}
			sendMessage(resp)
		case "log_url":
			var url string
			if err := json.Unmarshal(req.Payload, &url); err != nil {
				log.Printf("Error unmarshalling log_url payload: %v", err)
				continue
			}
			// Ignore logging the app's own GUI.
			if strings.HasPrefix(url, "http://127.0.0.1:58141") {
				continue
			}

			// Send the URL to the internal API
			go func(u string) {
				jsonData, _ := json.Marshal(map[string]string{"url": u})
				resp, err := http.Post(internalAPI+"/log-web-event", "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("Failed to send web event to internal API: %v", err)
					return
				}
				if err := resp.Body.Close(); err != nil {
					log.Printf("Failed to close response body: %v", err)
				}
			}(url)

		case "log_web_metadata":
			var payload WebMetadataPayload
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling log_web_metadata payload: %v", err)
				continue
			}
			go func(p WebMetadataPayload) {
				jsonData, _ := json.Marshal(p)
				resp, err := http.Post(internalAPI+"/log-web-metadata", "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("Failed to send web metadata to internal API: %v", err)
					return
				}
				if err := resp.Body.Close(); err != nil {
					log.Printf("Failed to close response body: %v", err)
				}
			}(payload)
		case "get_web_blocklist":
			list, err := data.LoadWebBlocklist()
			if err != nil {
				log.Printf("Error loading web blocklist: %v", err)
				continue
			}
			resp := Response{
				Type:    "web_blocklist",
				Payload: list,
			}
			sendMessage(resp)
		case "add_to_web_blocklist":
			var domain string
			if err := json.Unmarshal(req.Payload, &domain); err != nil {
				log.Printf("Error unmarshalling add_to_web_blocklist payload: %v", err)
				continue
			}
			if _, err := data.AddWebsiteToBlocklist(domain); err != nil {
				log.Printf("Error adding to web blocklist: %v", err)
			}
		default:
			// Optionally handle unknown message types.
		}
	}
}

// pollWebBlocklist periodically checks for changes in the web blocklist and sends updates to the extension.
func pollWebBlocklist() {
	log := data.GetLogger()
	var lastBlocklist []string
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := http.Get(internalAPI + "/get-web-blocklist")
		if err != nil {
			log.Printf("Failed to get web blocklist from internal API: %v", err)
			continue
		}

		var list []string
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			log.Printf("Failed to decode web blocklist from internal API: %v", err)
			if err := resp.Body.Close(); err != nil {
				log.Printf("Failed to close response body: %v", err)
			}
			continue
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}

		// Only send an update if the blocklist has changed.
		if !reflect.DeepEqual(list, lastBlocklist) {
			lastBlocklist = list
			resp := Response{
				Type:    "web_blocklist",
				Payload: list,
			}
			sendMessage(resp)
		}
	}
}

// sendMessage sends a message to the browser extension.
func sendMessage(resp Response) {
	log := data.GetLogger()
	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	// The native messaging protocol requires that the message length be sent first.
	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(b))); err != nil {
		log.Printf("Error writing message length: %v", err)
		return
	}

	// Then, the message body is sent.
	if _, err := os.Stdout.Write(b); err != nil {
		log.Printf("Error writing message body: %v", err)
		return
	}
}
