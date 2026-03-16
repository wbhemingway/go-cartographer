package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/wbhemingway/go-cartographer/internal/models"
)

// needs to handle errors more granually TODO
func (apiCfg *ApiConfig) handleRender(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	mapID, err := apiCfg.storeMapConfig(r.Context(), bytes)
	if err != nil {
		http.Error(w, "Internal server error during ingestion", http.StatusInternalServerError)
		return
	}

	err = apiCfg.renderWorld(r.Context(), mapID)
	if err != nil {
		http.Error(w, "Internal server error during rendering", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(models.MapResponse{
		ID: mapID,
	})
	if err != nil {
		slog.Error("Failed to write JSON response", "error", err)
	}
}

func (apiCfg *ApiConfig) storeMapConfig(ctx context.Context, rawJSON []byte) (string, error) {
	userID, ok := ctx.Value("UserID").(string)
	if !ok {
		slog.Error("Coud not find user id in context")
		return "", fmt.Errorf("failed to get user from context")
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
		Status:           "pending",
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
		//handle later
		return
	}

	userID, _ := r.Context().Value("UserID").(string)

	if mapData.CreatorID != userID {
		slog.Warn("Unauthorized access attempt",
			"requester", userID,
			"owner", mapData.CreatorID,
			"mapID", mapID,
		)
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}

	if mapData.Status != "completed" {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.MapResponse{
		ID:     mapID,
		Status: "completed",
		URL:    signedURL,
	}
	json.NewEncoder(w).Encode(response)
}
