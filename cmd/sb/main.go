package main

import (
	"log"

	"github.com/fogleman/gg"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

func main() {
	cfg := renderer.DefaultConfig()
	engine := renderer.New(cfg)

	world := renderer.World{
		Width:  2,
		Height: 2,
		Tiles: []renderer.Tile{
			{X: 0, Y: 0, Terrain: "grass"},
			{X: 1, Y: 0, Terrain: "grass", Structure: "tree", Creature: "goblin"},
			{X: 0, Y: 1, Terrain: "dirt", Creature: "goblin"},
			{X: 1, Y: 1, Terrain: "water"},
		},
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
