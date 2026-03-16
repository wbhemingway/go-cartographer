package main

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/wbhemingway/go-cartographer/internal/models"
)

//needs to specifically handle document not found, TODO
func getMapMetadata(ctx context.Context, fc *firestore.Client, mapID string) (models.MapMetadata, error) {
	docSnap, err := fc.Collection("maps").Doc(mapID).Get(ctx)
	if err != nil {
		slog.Error("Could not get map config from mapID", "error", err)
		return models.MapMetadata{}, err
	}

	var mapdata models.MapMetadata
	err = docSnap.DataTo(&mapdata)
	if err != nil {
		slog.Error("Could not convert firestore entry to MapMetadata", "error", err)
		return models.MapMetadata{}, err
	}

	return mapdata, nil
}
