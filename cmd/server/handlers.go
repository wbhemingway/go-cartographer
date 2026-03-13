package main

import (
	"encoding/json"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/wbhemingway/go-cartographer/internal/models"
)

func (apiCfg *ApiConfig) HandleRender(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("CARTOGRAPHER_API_KEY")
	if apiKey != "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}
	var world models.World
	err := json.NewDecoder(r.Body).Decode(&world)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	img, err := apiCfg.engine.Render(r.Context(), world)
	if err != nil {
		http.Error(w, "Failed to render map", http.StatusInternalServerError)
		log.Printf("Render error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	err = png.Encode(w, img)
	if err != nil {
		log.Printf("Failed to encode PNG response: %v", err)
	}
}
