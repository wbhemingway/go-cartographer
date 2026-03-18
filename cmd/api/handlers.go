package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/wbhemingway/go-cartographer/internal/models"
	"google.golang.org/api/iterator"
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

	msg := &pubsub.Message{Data: []byte(mapID)}
	result := apiCfg.pubsubPublisher.Publish(r.Context(), msg)
	_, err = result.Get(r.Context())
	if err != nil {
		slog.Error("Failed to publish render job to queue", "mapID", mapID, "error", err)
		http.Error(w, "failed to queue map for rendering", http.StatusInternalServerError)
		return
	}

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
	mapData, mapID, err := apiCfg.getAuthorizedMap(w, r)
	if err != nil {
		// getAuthorized map will have sent errors through writer
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

func (apiCfg *ApiConfig) handleDelMap(w http.ResponseWriter, r *http.Request) {
	_, mapID, err := apiCfg.getAuthorizedMap(w, r)
	if err != nil {
		// getAuthorized map will have sent errors through writer
		return
	}

	_, err = apiCfg.firestoreClient.Collection("maps").Doc(mapID).Delete(r.Context())
	if err != nil {
		slog.Error("Failed to delete map from database", "error", err)
		http.Error(w, "Failed to delete map from database", http.StatusInternalServerError)
		return
	}

	err = apiCfg.storageClient.Bucket(apiCfg.bucketName).Object("images/" + mapID + ".png").Delete(r.Context())
	if err != nil {
		slog.Error("Failed to delete map image from storage", "error", err)
	}

	err = apiCfg.storageClient.Bucket(apiCfg.bucketName).Object("configs/" + mapID + ".json").Delete(r.Context())
	if err != nil {
		slog.Error("Failed to delete map config from storage", "error", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (apiCfg *ApiConfig) handleListMaps(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(userIDKey).(string)
	iter := apiCfg.firestoreClient.Collection("maps").Where("creator_id", "==", userID).Documents(r.Context())
	defer iter.Stop()

	responses := make([]models.MapResponse, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			slog.Error("Failed to look through documents", "error", err)
			http.Error(w, "Failed to look through documents", http.StatusInternalServerError)
			return
		}
		var mapData models.MapMetadata
		err = doc.DataTo(&mapData)
		if err != nil {
			slog.Error("Failed to unmarshal document", "id", doc.Ref.ID, "error", err)
			continue
		}
		responses = append(responses, models.MapResponse{
			ID:     mapData.ID,
			Status: mapData.Status,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}
