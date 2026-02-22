package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/seanchuatech/go-webhook-dispatcher/internal/dispatcher"
	"github.com/seanchuatech/go-webhook-dispatcher/internal/handlers"
)

// Server wraps the standard library HTTP server
type Server struct {
	httpServer *http.Server
}

// New initializes a new HTTP server with sensible timeouts
func New(addr string) *Server {
	mux := http.NewServeMux()

	// Phase 3: Initialize dispatcher with a worker pool of 100 and a queue size of 10,000
	d := dispatcher.New(100, 10000)
	ingestHandler := handlers.NewIngestHandler(d)

	// A basic endpoint we can test
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	// The main ingestion endpoint
	mux.HandleFunc("/ingest", ingestHandler.HandleIngest)

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
			// Best practice: protect against slow client attacks
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}
}

// Start runs the HTTP server. It blocks until the server fails or is closed.
func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop attempts to gracefully shut down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
