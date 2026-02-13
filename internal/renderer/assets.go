package renderer

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
)

type AssetManager struct {
	basePath string
	cache    map[string]image.Image
	//TODO: mu map if speed becomes an issue
	mu sync.RWMutex
}

func NewAssetManager(basePath string, size int) *AssetManager {
	am := &AssetManager{
		basePath: basePath,
		cache:    make(map[string]image.Image),
	}
	am.cache["__missing__"] = generatePlaceholder(size)
	return am
}

func (am *AssetManager) Get(name string) (image.Image, error) {
	am.mu.RLock()
	img, ok := am.cache[name]
	am.mu.RUnlock()

	if ok {
		return img, nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	img, ok = am.cache[name]
	if ok {
		return img, nil
	}

	fullPath := filepath.Join(am.basePath, name+".png")
	file, err := os.Open(fullPath)
	if err != nil {
		return am.cache["__missing__"], fmt.Errorf("failed to open asset %s: %w", name, err)
	}

	defer file.Close()
	loadedImg, _, err := image.Decode(file)
	if err != nil {
		return am.cache["__missing__"], fmt.Errorf("failed to decode asset %s: %w", name, err)
	}

	am.cache[name] = loadedImg
	return loadedImg, nil
}

func generatePlaceholder(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	magenta := color.RGBA{255, 0, 255, 255}
	for x := range size {
		for y := range size {
			img.Set(x, y, magenta)
		}
	}
	return img
}
