package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/models"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

type pushRequest struct {
	Message struct {
		Data []byte `json:"data"`
		ID   string `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler).With("service", "cartographer-worker")
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
	workerCfg := &WorkerConfig{
		engine:          engine,
		storageClient:   storageClient,
		firestoreClient: firestoreClient,
		bucketName:      bucketName,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pubsub/push", workerCfg.handlePush)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Worker started, listening for Push messages", "port", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("Worker server crashed", "error", err)
		os.Exit(1)
	}
}

func (cfg *WorkerConfig) handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req pushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode push request", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	mapID := string(req.Message.Data)
	slog.Info("Received push render job", "mapID", mapID, "msgID", req.Message.ID)

	err := cfg.renderWorld(r.Context(), mapID)
	if err != nil {
		slog.Error("Render job failed", "mapID", mapID, "error", err)
		if errors.Is(err, models.ErrInvalidConfig) || errors.Is(err, models.ErrMapNotFound) {
			_, updateErr := cfg.firestoreClient.Collection("maps").Doc(mapID).Update(r.Context(), []firestore.Update{
				{Path: "status", Value: "failed"},
			})
			if updateErr != nil {
				slog.Error("Failed to set map status to failed", "mapID", mapID, "error", updateErr)
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, "Transient error, please retry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
