package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seanchuatech/go-webhook-dispatcher/internal/repository"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://dispatcher:secretpassword@localhost:5432/webhook_db?sslmode=disable"
	}

	// 2. Initialize the Database Repository
	repo, err := repository.NewEventRepository(dbUrl)
	if err != nil {
		slog.Error("failed to connect to database. Starting without persistence", "error", err)
		repo = nil // We can run the app without a DB if we want, gracefully degrading
	}

	// 3. Initialize the server
	addr := ":" + port
	srv := server.New(addr, repo)

	// 4. Start the server in a goroutine so it doesn't block
	go func() {
		slog.Info("starting server", "addr", addr)
		if err := srv.Start(); err != nil {
			slog.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a signal is received (managed by signal.NotifyContext)
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown
	stop()
	slog.Info("shutting down gracefully, press Ctrl+C again to force")

	// 6. Provide a timeout context for the shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	if repo != nil {
		_ = repo.Close()
	}

	slog.Info("server exited")
}
