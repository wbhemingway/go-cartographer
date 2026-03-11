package renderer

import (
	"testing"
)

func TestAssetManager_Get(t *testing.T) {
	cfg := DefaultConfig()
	tests := []struct {
		name           string
		assetPath      string
		expectFallback bool
	}{
		{
			name:           "Missing Asset Returns Magenta Placeholder",
			assetPath:      "missing/void",
			expectFallback: true,
		},
		{
			name:           "Empty String Returns Magenta Placeholder",
			assetPath:      "",
			expectFallback: true,
		},
		{
			name:           "Valid Asset Loads Actual Image (Not Magenta)",
			assetPath:      "terrain/grass",
			expectFallback: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := NewAssetManager("../../assets", cfg.TileSize)
			img, _ := am.Get(tt.assetPath)

			bounds := img.Bounds()
			if bounds.Dx() != cfg.TileSize || bounds.Dy() != cfg.TileSize {
				t.Errorf("Expected %dx%d, got %dx%d", cfg.TileSize, cfg.TileSize, bounds.Dx(), bounds.Dy())
			}

			r, g, b, _ := img.At(0, 0).RGBA()
			isMagenta := (r == 65535 && g == 0 && b == 65535)

			if tt.expectFallback && !isMagenta {
				t.Errorf("Expected fallback magenta image, got R:%d G:%d B:%d", r, g, b)
			}

			if !tt.expectFallback && isMagenta {
				t.Errorf("Expected valid image for %s, but got the magenta fallback", tt.assetPath)
			}
		})
	}
}
