package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Server wraps the standard library HTTP server
type Server struct {
	httpServer *http.Server
}

// New initializes a new HTTP server with sensible timeouts
func New(addr string) *Server {
	mux := http.NewServeMux()

	// A basic endpoint we can test
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

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
