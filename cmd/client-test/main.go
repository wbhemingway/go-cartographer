package main

import (
	"context"
	"encoding/json"
	"fmt"
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
			{X: 0, Y: 1, Terrain: "sandy", Creature: "goblin", Structure: "tree"},
			{X: 1, Y: 1, Terrain: "grass_flowers"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Sending POST request to queue the map render...")

	respBody, err := apiClient.RequestMap(ctx, world)
	if err != nil {
		log.Fatalf("Failed to queue map: %v", err)
	}
	defer respBody.Close()

	data, err := io.ReadAll(respBody)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var postRes models.MapResponse
	err = json.Unmarshal(data, &postRes)
	if err != nil {
		log.Fatalf("Failed to read POST response body: %v", err)
	}

	log.Printf("Success! Map queued with ID: %s. Status: %s", postRes.ID, postRes.Status)

	var finalURL string
	pollInterval := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Fatal("Timed out waiting for map to render")
		default:
		}

		log.Printf("Polling GET /maps/%s ...", postRes.ID)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/maps/%s", apiURL, postRes.ID), nil)
		if err != nil {
			log.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)

		getResp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("Failed to call GET endpoint: %v", err)
		}

		if getResp.StatusCode != http.StatusOK {
			log.Fatalf("GET endpoint returned unexpected status: %d", getResp.StatusCode)
		}

		getBody, _ := io.ReadAll(getResp.Body)
		getResp.Body.Close()

		var getRes models.MapResponse
		err = json.Unmarshal(getBody, &getRes)
		if err != nil {
			log.Fatalf("Failed to parse GET JSON: %v", err)
		}

		if getRes.Status == models.StatusCompleted {
			log.Println("Render complete!")
			finalURL = getRes.URL
			break
		} else if getRes.Status == "failed" {
			log.Fatal("Worker failed to render the map (Poison Pill). Check worker logs.")
		}

		log.Printf("Status is still '%s'. Waiting %v...", getRes.Status, pollInterval)
		time.Sleep(pollInterval)
	}

	log.Println("Downloading image from Signed URL...")
	imgResp, err := http.Get(finalURL)
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
