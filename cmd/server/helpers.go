package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

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
		return models.MapMetadata{}, err
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

func (apiCfg *ApiConfig) getAuthorizedMap(w http.ResponseWriter, r *http.Request) (models.MapMetadata, string, error) {
	mapID := r.PathValue("mapID")
	mapData, err := getMapMetadata(r.Context(), apiCfg.firestoreClient, mapID)
	if err != nil {
		if errors.Is(err, models.ErrMapNotFound) {
			http.Error(w, "Map not found", http.StatusNotFound)
			return models.MapMetadata{}, "", err
		}

		if errors.Is(err, models.ErrInvalidConfig) {
			slog.Error("Corrupted map config in database", "mapID", mapID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return models.MapMetadata{}, "", err
		}

		slog.Error("Failed to retrieve map", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return models.MapMetadata{}, "", err
	}

	userID, _ := r.Context().Value(userIDKey).(string)

	ok := userOwnsMapCheck(userID, mapData.CreatorID)
	if !ok {
		slog.Warn("Unauthorized access attempt",
			"requester", userID,
			"owner", mapData.CreatorID,
			"mapID", mapID,
		)
		http.Error(w, "Access Denied", http.StatusForbidden)
		return models.MapMetadata{}, "", models.ErrUnauthorized
	}

	return mapData, mapID, nil
}
