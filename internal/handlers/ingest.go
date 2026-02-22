package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/seanchuatech/go-webhook-dispatcher/internal/dispatcher"
	"github.com/seanchuatech/go-webhook-dispatcher/internal/domain"
)

// IngestHandler handles receiving incoming events
type IngestHandler struct {
	dispatcher *dispatcher.Dispatcher
}

// NewIngestHandler creates a new handler
func NewIngestHandler(d *dispatcher.Dispatcher) *IngestHandler {
	return &IngestHandler{
		dispatcher: d,
	}
}

// HandleIngest processes incoming POST requests containing events.
func (h *IngestHandler) HandleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse the incoming JSON payload into our Event struct
	var event domain.Event
	// Best Practice: Prevent massive payloads from crashing the server by limiting request body size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		slog.Error("failed to decode request body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if event.DestinationURL == "" {
		http.Error(w, "destination_url is required", http.StatusBadRequest)
		return
	}

	// Log receipt of the event
	slog.Info("received event", "event_id", event.ID, "destination", event.DestinationURL)

	// 2. Dispatch the event (Synchronous for Phase 2)
	// We pass the request context so if the client disconnects, we can cancel the dispatch
	err := h.dispatcher.Dispatch(r.Context(), event)
	if err != nil {
		slog.Error("failed to dispatch event", "event_id", event.ID, "error", err)
		http.Error(w, "Failed to dispatch event", http.StatusInternalServerError)
		return
	}

	// 3. Return success
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}
