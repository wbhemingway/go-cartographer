package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/wbhemingway/go-cartographer/internal/models"
)

func (apiCfg *ApiConfig) handleRender(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	mapID, err := apiCfg.storeMapConfig(r.Context(), bytes)
	if err != nil {
		if errors.Is(err, models.ErrUnauthorized) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, models.ErrInvalidConfig) {
			http.Error(w, "Invalid map configuration", http.StatusBadRequest)
			return
		}

		slog.Error("Database error during ingestion", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	go func(id string) {
		err := apiCfg.renderWorld(context.Background(), id)
		if err != nil {
			// TODO update document to failed rather than pending
			slog.Error("Background rendering failed", "mapID", id, "error", err)
		}
	}(mapID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(models.MapResponse{
		ID:     mapID,
		Status: models.StatusPending,
	})
	if err != nil {
		slog.Error("Failed to write JSON response", "error", err)
	}
}

func (apiCfg *ApiConfig) storeMapConfig(ctx context.Context, rawJSON []byte) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		slog.Error("Coud not find user id in context")
		return "", models.ErrUnauthorized
	}

	var validationWorld models.World

	err := json.Unmarshal(rawJSON, &validationWorld)
	if err != nil {
		slog.Warn("Failed to unmarshal user map config", "error", err)
		return "", models.ErrInvalidConfig
	}

	mapID, err := uuid.NewV7()
	if err != nil {
		slog.Error("Failed to generate ID", "error", err)
		return "", err
	}

	objectName := "configs/" + mapID.String() + ".json"

	obj := apiCfg.storageClient.Bucket(apiCfg.bucketName).Object(objectName)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	writer.ContentType = "application/json"
	_, err = writer.Write(rawJSON)
	if err != nil {
		slog.Error("Failed to write to bucket", "error", err, "objectName", objectName)
		return "", err

	}
	err = writer.Close()
	if err != nil {
		slog.Error("Error closing writer", "error", err)
		return "", err
	}
	mapData := models.MapMetadata{
		ID:               mapID.String(),
		CreatorID:        userID,
		ConfigObjectName: objectName,
		CreatedAt:        time.Now().UTC(),
		Status:           models.StatusPending,
	}

	_, err = apiCfg.firestoreClient.Collection("maps").Doc(mapID.String()).Set(ctx, mapData)
	if err != nil {
		slog.Error("Failed to write to firestore", "error", err)
		return "", err
	}

	return mapID.String(), nil
}

func (apiCfg *ApiConfig) handleGetMap(w http.ResponseWriter, r *http.Request) {
	mapID := r.PathValue("mapID")
	mapData, err := getMapMetadata(r.Context(), apiCfg.firestoreClient, mapID)
	if err != nil {
		if errors.Is(err, models.ErrMapNotFound) {
			http.Error(w, "Map not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, models.ErrInvalidConfig) {
			slog.Error("Corrupted map config in database", "mapID", mapID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		slog.Error("Failed to retrieve map", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
		return
	}

	if mapData.Status != models.StatusCompleted {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.MapResponse{
			ID:     mapID,
			Status: mapData.Status,
		})
		return
	}

	objectName := "images/" + mapID + ".png"
	signedURL, err := storage.SignedURL(apiCfg.bucketName, objectName, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		slog.Error("Failed to generate signed URL", "error", err)
		http.Error(w, "Failed to generate image link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=900")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.MapResponse{
		ID:     mapID,
		Status: models.StatusCompleted,
		URL:    signedURL,
	}
	json.NewEncoder(w).Encode(response)
}
