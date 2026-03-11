package main

import (
	"encoding/json"
	"image/png"
	"log"
	"net/http"

	"github.com/wbhemingway/go-cartographer/internal/models"
)

func (s *Server) HandleRender(w http.ResponseWriter, r *http.Request) {
	var world models.World
	err := json.NewDecoder(r.Body).Decode(&world)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	img, err := s.engine.Render(r.Context(), world)
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
