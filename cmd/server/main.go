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

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler).With("service", "cartographer-api")
	slog.SetDefault(logger)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
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

	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		slog.Error("Failed to create firestore client", "error", err)
		os.Exit(1)
	}
	defer firestoreClient.Close()

	cfg := renderer.DefaultConfig()
	engine := renderer.New(cfg)
	apiCfg := &ApiConfig{
		engine:          engine,
		storageClient:   storageClient,
		firestoreClient: firestoreClient,
		bucketName:      bucketName,
	}

	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /maps", apiCfg.authMiddleware(apiCfg.handleRender))
	mux.HandleFunc("GET /maps/{mapID}", apiCfg.authMiddleware(apiCfg.handleGetMap))
	mux.HandleFunc("DELETE /maps/{mapID}", apiCfg.authMiddleware(apiCfg.handleDelMap))
	mux.HandleFunc("GET /maps", apiCfg.authMiddleware(apiCfg.handleListMaps))

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
