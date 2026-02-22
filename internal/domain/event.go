package domain

import (
	"encoding/json"
	"time"
)

// Event represents the core structure of a webhook payload that we ingest.
type Event struct {
	ID             string          `json:"id"`
	Source         string          `json:"source"`
	Type           string          `json:"type"`
	DestinationURL string          `json:"destination_url"`
	Payload        json.RawMessage `json:"payload"` // Flexible payload block
	CreatedAt      time.Time       `json:"created_at"`
}
