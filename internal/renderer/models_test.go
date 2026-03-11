package renderer

import (
	"encoding/json"
	"testing"
)

func TestWorld_JSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		expectErr bool
		validate  func(t *testing.T, w World)
	}{
		{
			name:      "Valid Full JSON Payload",
			expectErr: false,
			payload: `{
				"width": 2,
				"height": 2,
				"tiles": [
					{"x": 0, "y": 0, "terrain": "grass", "structure": "wall", "creature": "goblin"}
				]
			}`,
			validate: func(t *testing.T, w World) {
				if w.Width != 2 || w.Height != 2 {
					t.Errorf("expected 2x2, got %dx%d", w.Width, w.Height)
				}
				if len(w.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(w.Tiles))
				}
				if w.Tiles[0].Terrain != "grass" {
					t.Errorf("expected terrain 'grass', got '%s'", w.Tiles[0].Terrain)
				}
			},
		},
		{
			name:      "Malformed JSON (Syntax Error)",
			expectErr: true,
			payload: `{
				"width": 2,
				"height": 2,
				"tiles": [
					{"x": 0, "y": 0, "terrain": "grass" // Missing closing brackets
			}`,
			validate: nil,
		},
		{
			name:      "Wrong Data Types (String instead of Int)",
			expectErr: true,
			payload: `{
				"width": "two",
				"height": "two",
				"tiles": []
			}`,
			validate: nil,
		},
		{
			name:      "Empty JSON Object",
			expectErr: false,
			payload:   `{}`,
			validate: func(t *testing.T, w World) {
				if w.Width != 0 || w.Height != 0 {
					t.Errorf("expected defaults 0x0, got %dx%d", w.Width, w.Height)
				}
				if len(w.Tiles) != 0 {
					t.Errorf("expected 0 tiles, got %d", len(w.Tiles))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w World
			err := json.Unmarshal([]byte(tt.payload), &w)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if !tt.expectErr && tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}
