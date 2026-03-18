package main

import (
	"context"
	"encoding/json"
	"errors"
	"image/png"
	"io"
	"log/slog"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (workerCfg *WorkerConfig) renderWorld(ctx context.Context, mapID string) error {
	docSnap, err := workerCfg.firestoreClient.Collection("maps").Doc(mapID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			slog.Error("Map document does not exist", "mapID", mapID)
			return models.ErrMapNotFound
		}
		slog.Error("Network error getting map config", "error", err)
		return err
	}

	var mapdata models.MapMetadata
	err = docSnap.DataTo(&mapdata)
	if err != nil {
		slog.Error("Could not convert firestore entry to MapMetadata", "error", err)
		return models.ErrInvalidConfig
	}

	confObj := workerCfg.storageClient.Bucket(workerCfg.bucketName).Object(mapdata.ConfigObjectName)
	reader, err := confObj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			slog.Error("Config file missing from storage", "object", mapdata.ConfigObjectName)
			return models.ErrInvalidConfig
		}
		slog.Error("Network error reading config from storage", "error", err)
		return err
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		slog.Error("Could not read all from the confObj", "error", err)
		return err
	}

	var world models.World
	err = json.Unmarshal(body, &world)
	if err != nil {
		slog.Error("Could not unmarshal body into world struct", "error", err)
		return models.ErrInvalidConfig
	}

	img, err := workerCfg.engine.Render(ctx, world)
	if err != nil {
		slog.Error("Render error", "error", err)
		return models.ErrInvalidConfig
	}

	objectName := "images/" + mapID + ".png"
	bucket := workerCfg.storageClient.Bucket(workerCfg.bucketName)
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(ctx)
	writer.ContentType = "image/png"
	defer writer.Close()

	err = png.Encode(writer, img)
	if err != nil {
		slog.Error("Failed to write image to storage", "error", err)
		return err
	}

	err = writer.Close()
	if err != nil {
		slog.Error("Failed to finalize image upload", "error", err)
		return err
	}

	_, err = workerCfg.firestoreClient.Collection("maps").Doc(mapID).Update(ctx, []firestore.Update{
		{Path: "status", Value: models.StatusCompleted},
	})
	if err != nil {
		slog.Error("Failed to update the database entry to completed", "error", err)
		return err
	}

	return nil
}
