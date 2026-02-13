package renderer

import (
	"image"
	"log"
	"runtime"
	"sync"

	"github.com/fogleman/gg"
)

type Config struct {
	TileSize  int
	AssetPath string
}

type Engine struct {
	cfg    Config
	assets *AssetManager
}

func DefaultConfig() Config {
	return Config{
		TileSize:  64,
		AssetPath: "./assets",
	}
}

type result struct {
	x, y int
	tile image.Image
}

func New(cfg Config) *Engine {
	return &Engine{
		cfg: cfg,
		//TODO: Validate
		assets: NewAssetManager(cfg.AssetPath, cfg.TileSize),
	}
}

func (e *Engine) Render(w World) (image.Image, error) {
	dcWidth := w.Width * e.cfg.TileSize
	dcHeight := w.Height * e.cfg.TileSize
	dc := gg.NewContext(dcWidth, dcHeight)

	numWorkers := runtime.NumCPU() * 4
	jobs := make(chan Tile, len(w.Tiles))
	results := make(chan result, len(w.Tiles))
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				patch := e.renderTile(t)
				results <- result{x: t.X, y: t.Y, tile: patch}
			}
		}()
	}

	for _, t := range w.Tiles {
		jobs <- t
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	for res := range results {
		posX := float64(res.x * e.cfg.TileSize)
		posY := float64(res.y * e.cfg.TileSize)
		dc.DrawImage(res.tile, int(posX), int(posY))
	}

	return dc.Image(), nil
}

func (e *Engine) renderTile(t Tile) image.Image {
	tdc := gg.NewContext(e.cfg.TileSize, e.cfg.TileSize)

	img, err := e.assets.Get("terrain/" + t.Terrain)
	if err != nil {
		log.Println(err)
	}
	tdc.DrawImage(img, 0, 0)

	if t.Structure != "" {
		img, err := e.assets.Get("structure/" + t.Structure)
		if err == nil {
			tdc.DrawImage(img, 0, 0)
		} else {
			log.Println(err)
		}
	}

	if t.Creature != "" {
		img, err := e.assets.Get("creature/" + t.Creature)
		if err == nil {
			tdc.DrawImage(img, 0, 0)
		} else {
			log.Println(err)
		}
	}

	return tdc.Image()
}
