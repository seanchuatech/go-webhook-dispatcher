package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seanchuatech/go-webhook-dispatcher/internal/server"
)

func main() {
	// Phase 1: Introduce structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Phase 1: Implement graceful shutdown
	// Listen for SIGINT (Ctrl+C) and SIGTERM (Kubernetes/Docker shutdown)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := ":8080"
	srv := server.New(addr)

	// Phase 1: Build basic HTTP server
	// Run server in a goroutine so it doesn't block
	go func() {
		slog.Info("starting server", "addr", addr)
		if err := srv.Start(); err != nil {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for the interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown
	stop()
	slog.Info("shutting down gracefully, press Ctrl+C again to force")

	// Context with timeout to give active connections time to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited")
}
