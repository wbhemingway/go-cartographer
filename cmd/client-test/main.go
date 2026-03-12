package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/wbhemingway/go-cartographer/internal/models"
	"github.com/wbhemingway/go-cartographer/pkg/client"
)

func main() {
	apiClient := client.New("http://localhost:8080")

	world := models.World{
		Width:  2,
		Height: 2,
		Tiles: []models.Tile{
			{X: 0, Y: 0, Terrain: "grass", Structure: "hut"},
			{X: 1, Y: 0, Terrain: "water"},
			{X: 0, Y: 1, Terrain: "sand", Creature: "goblin"},
			{X: 1, Y: 1, Terrain: "grass"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Sending map request to server...")

	imageStream, err := apiClient.RequestMap(ctx, world)
	if err != nil {
		log.Fatalf("Failed to get map from server: %v", err)
	}
	defer imageStream.Close()

	outputFile, err := os.Create("client_test_output.png")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	bytesWritten, err := io.Copy(outputFile, imageStream)
	if err != nil {
		log.Fatalf("Failed to save image: %v", err)
	}

	log.Printf("Success! Saved map to client_test_output.png (%d bytes)", bytesWritten)
}
