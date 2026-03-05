package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

func main() {
	cfg := renderer.DefaultConfig()
	engine := renderer.New(cfg)
	world_file, err := os.ReadFile("sandbox_map.json")
	if err != nil {
		log.Fatalf("Failed to read map file: %v", err)
	}
	var world renderer.World
	err = json.Unmarshal(world_file, &world)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Println("Rendering map...")
	img, err := engine.Render(world)
	if err != nil {
		log.Fatalf("Render failed: %v", err)
	}

	outputPath := "sandbox_map.png"
	if err := gg.SavePNG(outputPath, img); err != nil {
		log.Fatalf("Failed to save image: %v", err)
	}

	log.Printf("Success! Map saved to %s", outputPath)
}
