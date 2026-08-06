package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	draftapi "github.com/coopa11y/DraftMeld/backend/internal/api"
	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	draftsqlite "github.com/coopa11y/DraftMeld/backend/internal/persistence/sqlite"
)

var version = "0.2.0-dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	address := envOrDefault("DRAFTMELD_ADDRESS", ":8080")
	dataDirectory := envOrDefault("DRAFTMELD_DATA_DIR", "./data")
	if err := os.MkdirAll(dataDirectory, 0o750); err != nil {
		logger.Error("create data directory", "error", err)
		os.Exit(1)
	}
	store, err := draftsqlite.Open(filepath.Join(dataDirectory, "draftmeld.db"))
	if err != nil {
		logger.Error("open persistence", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	draftService, err := application.NewDraftService(store, draft.DemoCatalog(), application.DemoLeagueConfiguration())
	if err != nil {
		logger.Error("configure draft service", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              address,
		Handler:           draftapi.NewRouter(logger, version, draftService),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("DraftMeld started", "address", address, "version", version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
