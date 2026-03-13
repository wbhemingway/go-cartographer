package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler).With("service", "cartographer-api")
	slog.SetDefault(logger)

	bucketName := os.Getenv("CARTOGRAPHER_BUCKET_NAME")
	if bucketName == "" {
		slog.Error("CARTOGRAPHER_BUCKET_NAME environment variable is not set")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		slog.Error("Failed to create storage client", "error", err)
		os.Exit(1)
	}
	defer storageClient.Close()

	cfg := renderer.DefaultConfig()
	engine := renderer.New(cfg)
	apiCfg := &ApiConfig{
		engine:        engine,
		storageClient: storageClient,
		bucketName:    bucketName,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /render", authMiddleware(apiCfg.HandleRender))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serve := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("Starting map rendering server", "addr", serve.Addr)
		if err := serve.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server listen error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutdown signal received, draining ongoing requests...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = serve.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Cartographer Server stopped cleanly")
}
