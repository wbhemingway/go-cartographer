package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/models"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

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

	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		slog.Error("Failed to create pubsub client", "error", err)
		os.Exit(1)
	}
	defer pubsubClient.Close()

	pubsubSubscriber := pubsubClient.Subscriber("map-render-jobs-sub")

	cfg := renderer.DefaultConfig()
	engine := renderer.New(cfg)
	workerCfg := &WorkerConfig{
		engine:           engine,
		storageClient:    storageClient,
		firestoreClient:  firestoreClient,
		pubsubSubscriber: pubsubSubscriber,
		bucketName:       bucketName,
	}

	slog.Info("Worker started, listening for map render jobs...")

	err = workerCfg.pubsubSubscriber.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		mapID := string(msg.Data)
		slog.Info("Received render job", "mapID", mapID)

		err := workerCfg.renderWorld(ctx, mapID)
		if err != nil {
			slog.Error("Could not render the json", "mapID", mapID, "error", err)

			if errors.Is(err, models.ErrInvalidConfig) || errors.Is(err, models.ErrMapNotFound) {
				_, updateErr := workerCfg.firestoreClient.Collection("maps").Doc(mapID).Update(ctx, []firestore.Update{
					{Path: "status", Value: "failed"},
				})
				if updateErr != nil {
					slog.Error("Failed to set map status to failed", "mapID", mapID, "error", updateErr)
				}
				msg.Ack()
				return
			}
			msg.Nack()
			return
		}
		msg.Ack()
	})
	if err != nil {
		slog.Error("Worker subscriber loop failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Worker shut down cleanly")
}
