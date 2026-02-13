package renderer

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

func New(cfg Config) *Engine {
	return &Engine{
		cfg: cfg,
		//TODO: Validate
		assets: NewAssetManager(cfg.AssetPath, cfg.TileSize),
	}
}
