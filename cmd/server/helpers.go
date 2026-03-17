package main

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/wbhemingway/go-cartographer/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getMapMetadata(ctx context.Context, fc *firestore.Client, mapID string) (models.MapMetadata, error) {
	docSnap, err := fc.Collection("maps").Doc(mapID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.MapMetadata{}, models.ErrMapNotFound
		}
		
		slog.Error("Could not get map config from mapID", "error", err)
		return models.MapMetadata{}, models.ErrMapNotFound
	}

	var mapdata models.MapMetadata
	err = docSnap.DataTo(&mapdata)
	if err != nil {
		slog.Error("Could not convert firestore entry to MapMetadata", "error", err)
		return models.MapMetadata{}, models.ErrInvalidConfig
	}

	return mapdata, nil
}

func userOwnsMapCheck(userID, ownerID string) bool {
	return userID == ownerID
}
