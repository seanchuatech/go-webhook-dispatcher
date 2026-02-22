package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/seanchuatech/go-webhook-dispatcher/internal/domain"
)

type EventRepository struct {
	db *sql.DB
}

// NewEventRepository initializes the connection to the database.
func NewEventRepository(connURL string) (*EventRepository, error) {
	db, err := sql.Open("postgres", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to the database")
	return &EventRepository{db: db}, nil
}

// Close closes the database connection.
func (r *EventRepository) Close() error {
	return r.db.Close()
}

// InsertEvent saves a new event with a given status.
func (r *EventRepository) InsertEvent(ctx context.Context, event domain.Event, status string) error {
	query := `
		INSERT INTO events (id, source, type, destination_url, payload, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, event.ID, event.Source, event.Type, event.DestinationURL, string(payloadJSON), status)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

// UpdateEventStatus updates the status of an existing event.
func (r *EventRepository) UpdateEventStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE events SET status = $1, updated_at = NOW() WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}
	return nil
}
