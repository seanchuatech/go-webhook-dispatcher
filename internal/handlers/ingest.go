package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/seanchuatech/go-webhook-dispatcher/internal/dispatcher"
	"github.com/seanchuatech/go-webhook-dispatcher/internal/domain"
	"github.com/seanchuatech/go-webhook-dispatcher/internal/repository"
)

// IngestHandler handles receiving incoming events
type IngestHandler struct {
	dispatcher *dispatcher.Dispatcher
	repo       *repository.EventRepository
}

// NewIngestHandler creates a new handler
func NewIngestHandler(d *dispatcher.Dispatcher, repo *repository.EventRepository) *IngestHandler {
	return &IngestHandler{
		dispatcher: d,
		repo:       repo,
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

	if event.Type == "" {
		http.Error(w, `{"error": "type is required"}`, http.StatusBadRequest)
		return
	}

	// Log receipt of the event
	slog.Info("received event", "event_id", event.ID, "destination", event.DestinationURL)

	// Insert the event as PENDING before queuing it
	if h.repo != nil {
		if err := h.repo.InsertEvent(r.Context(), event, "PENDING"); err != nil {
			slog.Error("failed to insert event into database", "error", err, "event_id", event.ID)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	// 2. Dispatch the event (Synchronous for Phase 2)
	// We pass the request context so if the client disconnects, we can cancel the dispatch
	err := h.dispatcher.Dispatch(r.Context(), event)
	if err != nil {
		slog.Error("dispatcher queue full", "error", err)
		http.Error(w, `{"error": "server is overloaded, please try again later"}`, http.StatusServiceUnavailable)
		return
	}

	// 3. Return success
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}
