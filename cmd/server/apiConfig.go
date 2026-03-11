package main

import "github.com/wbhemingway/go-cartographer/internal/renderer"

type ApiConfig struct {
	engine *renderer.Engine
}

func NewApiConfig(engine *renderer.Engine) *ApiConfig {
	return &ApiConfig{engine: engine}
}
