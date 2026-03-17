package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/wbhemingway/go-cartographer/internal/models"
	"github.com/wbhemingway/go-cartographer/pkg/client"
)

func main() {

	_ = godotenv.Load()
	apiURL := os.Getenv("CARTOGRAPHER_API_URL")
	apiKey := os.Getenv("CARTOGRAPHER_API_KEY")
	if apiURL == "" {
		log.Fatal("Fatal: CARTOGRAPHER_API_URL environment variable is not set")
	}
	apiClient := client.New(apiURL, apiKey, http.DefaultClient)

	world := models.World{
		Width:  2,
		Height: 2,
		Tiles: []models.Tile{
			{X: 0, Y: 0, Terrain: "grass", Structure: "hut"},
			{X: 1, Y: 0, Terrain: "water"},
			{X: 0, Y: 1, Terrain: "sandy", Creature: "goblin"},
			{X: 1, Y: 1, Terrain: "grass_flowers"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Sending map request to server...")

	respBody, err := apiClient.RequestMap(ctx, world)
	if err != nil {
		log.Fatalf("Failed to get response from server: %v", err)
	}
	defer respBody.Close()

	data, err := io.ReadAll(respBody)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var res models.MapResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Success! Map ID: %s", res.ID)
	log.Printf("Signed URL (valid for 15m): %s", res.URL)

	log.Println("Downloading image from Signed URL...")
	imgResp, err := http.Get(res.URL)
	if err != nil || imgResp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to download image from signed URL: %v", err)
	}
	defer imgResp.Body.Close()

	outFile, err := os.Create("test-render.png")
	if err != nil {
		log.Fatalf("Failed to create local file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, imgResp.Body)
	if err != nil {
		log.Fatalf("Failed to save image: %v", err)
	}

	log.Println("Verified! Image saved to test-render.png")
}
