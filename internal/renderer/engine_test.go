package renderer

import (
	"context"
	"testing"
	
	"github.com/wbhemingway/go-cartographer/internal/models"
)

func TestEngine_Render(t *testing.T) {
	cfg := DefaultConfig()
	engine := New(cfg)

	tests := []struct {
		name           string
		world          models.World
		expectErr      bool
		expectedWidth  int
		expectedHeight int
	}{
		{
			name: "Standard 2x2 World",
			world: models.World{
				Width: 2, Height: 2,
				Tiles: []models.Tile{
					{X: 0, Y: 0, Terrain: "grass"},
					{X: 1, Y: 1, Terrain: "missing_water"},
				},
			},
			expectErr:      false,
			expectedWidth:  128,
			expectedHeight: 128,
		},
		{
			name: "Empty World (0x0)",
			world: models.World{
				Width: 0, Height: 0,
				Tiles: []models.Tile{},
			},
			expectErr:      false,
			expectedWidth:  0,
			expectedHeight: 0,
		},
		{
			name: "Heavily Layered Single Tile",
			world: models.World{
				Width: 1, Height: 1,
				Tiles: []models.Tile{
					{X: 0, Y: 0, Terrain: "dirt", Structure: "wall", Creature: "goblin"},
				},
			},
			expectErr:      false,
			expectedWidth:  64,
			expectedHeight: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := engine.Render(context.Background(), tt.world)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if !tt.expectErr {
				bounds := img.Bounds()
				if bounds.Dx() != tt.expectedWidth || bounds.Dy() != tt.expectedHeight {
					t.Errorf("expected dimensions %dx%d, got %dx%d",
						tt.expectedWidth, tt.expectedHeight, bounds.Dx(), bounds.Dy())
				}
			}
		})
	}
}

func TestEngine_Render_ContextCancellation(t *testing.T) {
	cfg := DefaultConfig()
	engine := New(cfg)

	world := models.World{Width: 100, Height: 100, Tiles: make([]models.Tile, 10000)}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := engine.Render(ctx, world)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}
