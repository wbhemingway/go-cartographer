package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/wbhemingway/go-cartographer/internal/models"
)

func (apiCfg *ApiConfig) HandleRender(w http.ResponseWriter, r *http.Request) {
	var world models.World
	err := json.NewDecoder(r.Body).Decode(&world)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		slog.Info("Invalid JSON payload", "error", err)
		return
	}
	defer r.Body.Close()

	img, err := apiCfg.engine.Render(r.Context(), world)
	if err != nil {
		http.Error(w, "Failed to render map", http.StatusInternalServerError)
		slog.Error("Render error", "error", err)
		return
	}

	mapID, err := uuid.NewV7()
	if err != nil {
		http.Error(w, "Failed to generate ID", http.StatusInternalServerError)
		slog.Error("Failed to generate ID", "error", err)
		return
	}

	objectName := fmt.Sprintf("map-%s.png", mapID.String())
	bucket := apiCfg.storageClient.Bucket(apiCfg.bucketName)
	obj := bucket.Object(objectName)

	writer := obj.NewWriter(r.Context())
	writer.ContentType = "image/png"
	defer writer.Close()

	err = png.Encode(writer, img)
	if err != nil {
		http.Error(w, "Failed to write image to storage", http.StatusInternalServerError)
		slog.Error("Failed to write image to storage", "error", err)
		return
	}

	err = writer.Close()
	if err != nil {
		http.Error(w, "Failed to finalize image upload", http.StatusInternalServerError)
		slog.Error("Failed to finalize image upload", "error", err)
		return
	}

	signedURL, err := bucket.SignedURL(objectName, &storage.SignedURLOptions{
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
	response := models.RenderResponse{
		ID:  mapID.String(),
		URL: signedURL,
	}
	json.NewEncoder(w).Encode(response)
}
