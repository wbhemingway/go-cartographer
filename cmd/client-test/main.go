package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	cartographer "github.com/wbhemingway/go-cartographer/pkg/client"
)

func main() {
	_ = godotenv.Load()
	apiURL := os.Getenv("CARTOGRAPHER_API_URL")
	apiKey := os.Getenv("CARTOGRAPHER_API_KEY")

	if apiURL == "" {
		log.Fatal("Fatal: CARTOGRAPHER_API_URL environment variable is not set")
	}
	c := cartographer.NewClient(
		apiURL,
		cartographer.WithTimeout(10*time.Second),
		cartographer.WithAPIKey(apiKey),
	)

	ctx := context.Background()

	req := cartographer.WorldRequest{
		Width:  2,
		Height: 2,
		Tiles: []cartographer.Tile{
			{X: 0, Y: 0, Terrain: "grass"},
			{X: 1, Y: 1, Terrain: "water"},
		},
	}

	fmt.Println("Submitting render job...")
	resp, err := c.Render(ctx, req)
	if err != nil {
		log.Fatalf("Render failed: %v", err)
	}
	fmt.Printf("Job submitted! ID: %s, Status: %s\n", resp.ID, resp.Status)

	for {
		time.Sleep(2 * time.Second)
		statusResp, err := c.GetMap(ctx, resp.ID)
		if err != nil {
			log.Fatalf("Failed to fetch map status: %v", err)
		}

		fmt.Printf("Polling status: %s\n", statusResp.Status)
		if statusResp.Status == "completed" {
			fmt.Printf("Render successful! Image URL: %s\n", statusResp.URL)
			break
		}
	}

	mapList, err := c.ListMaps(ctx)
	if err != nil {
		log.Fatalf("Could not get the list of maps: %v", err)
	}

	for _, item := range mapList {
		if item.ID == resp.ID {
			fmt.Printf("The map with ID %v is in the list!\n", item.ID)
			break
		}
	}

	err = c.DeleteMap(ctx, resp.ID)
	if err != nil {
		log.Fatalf("There was an issue deleting map with id %v: %v\n", resp.ID, err)
	}

	_, err = c.GetMap(ctx, resp.ID)
	if err == nil {
		log.Fatalf("There was an issue deleting map with id %v: %v\n", resp.ID, err)
	}

	fmt.Println("Map successfully deleted")
}
