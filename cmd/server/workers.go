package main

import (
	"context"
	"encoding/json"
	"image/png"
	"io"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/wbhemingway/go-cartographer/internal/models"
)

func (apiCfg *ApiConfig) renderWorld(ctx context.Context, mapID string) error {
	docSnap, err := apiCfg.firestoreClient.Collection("maps").Doc(mapID).Get(ctx)
	if err != nil {
		slog.Error("Could not get map config from mapID", "error", err)
		return err
	}

	var mapdata models.MapMetadata
	err = docSnap.DataTo(&mapdata)
	if err != nil {
		slog.Error("Could not convert firestore entry to MapMetadata", "error", err)
		return err
	}

	confObj := apiCfg.storageClient.Bucket(apiCfg.bucketName).Object(mapdata.ConfigObjectName)
	reader, err := confObj.NewReader(ctx)
	if err != nil {
		slog.Error("Could not make reader for confObj.NewReader(ctx)", "error", err)
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
		return err
	}

	img, err := apiCfg.engine.Render(ctx, world)
	if err != nil {
		slog.Error("Render error", "error", err)
		return err
	}

	objectName := "images/" + mapID + ".png"
	bucket := apiCfg.storageClient.Bucket(apiCfg.bucketName)
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

	_, err = apiCfg.firestoreClient.Collection("maps").Doc(mapID).Update(ctx, []firestore.Update{
		{Path: "status", Value: "completed"},
	})
	if err != nil {
		slog.Error("Failed to update the database entry to completed", "error", err)
		return err
	}

	return nil
}
