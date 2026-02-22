package dispatcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/seanchuatech/go-webhook-dispatcher/internal/domain"
)

// Dispatcher handles sending events to their destination.
type Dispatcher struct {
	client      *http.Client
	eventStream chan domain.Event
}

// New creates a new Dispatcher with sensible HTTP client timeouts and starts the worker pool.
func New(workerCount int, maxQueueSize int) *Dispatcher {
	d := &Dispatcher{
		client: &http.Client{
			// Best Practice: Never use the default HTTP client in production as it has no timeout.
			Timeout: 10 * time.Second,
		},
		eventStream: make(chan domain.Event, maxQueueSize),
	}

	// Phase 3: Start the worker pool
	for i := 1; i <= workerCount; i++ {
		go d.worker(i)
	}

	return d
}

// worker listens on the eventStream channel and dispatches events.
func (d *Dispatcher) worker(id int) {
	for event := range d.eventStream {
		maxRetries := 5
		baseBackoff := 1 * time.Second

		for attempt := 1; attempt <= maxRetries; attempt++ {
			// Use a background context with timeout for the actual sending
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

			slog.Debug("worker processing event", "worker_id", id, "event_id", event.ID, "attempt", attempt)
			err := d.send(ctx, event)
			cancel() // Prevent context leak

			if err == nil {
				slog.Info("worker successfully dispatched event", "worker_id", id, "event_id", event.ID, "attempt", attempt)
				break // Success! Exit the retry loop
			}

			slog.Error("worker failed to send event", "worker_id", id, "event_id", event.ID, "attempt", attempt, "error", err)

			if attempt == maxRetries {
				slog.Error("max retries reached, event failed permanently", "worker_id", id, "event_id", event.ID)
				// Phase 5 will introduce Dead Letter Queue here
				break
			}

			// Exponential Backoff: Wait longer between each retry (1s, 2s, 4s, 8s, etc.)
			// In production, we would also add "jitter" (randomness) to prevent the Thundering Herd problem
			sleepDuration := baseBackoff * time.Duration(1<<(attempt-1))
			slog.Info("backing off before retry", "event_id", event.ID, "sleep_duration", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}
}

// Dispatch pushes the event to the channel. This handles the ingestion asynchronously.
func (d *Dispatcher) Dispatch(ctx context.Context, event domain.Event) error {
	select {
	case d.eventStream <- event:
		return nil
	default:
		// The channel is full! This means our workers are overwhelmed and the queue is backed up.
		return fmt.Errorf("dispatcher queue is full, dropping event %s", event.ID)
	}
}

// send executes the actual HTTP request.
func (d *Dispatcher) send(ctx context.Context, event domain.Event) error {
	// Create the HTTP request, forwarding only the payload portion
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, event.DestinationURL, bytes.NewBuffer(event.Payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
