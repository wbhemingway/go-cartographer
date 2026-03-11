package main

import "github.com/wbhemingway/go-cartographer/internal/renderer"

type Server struct {
	engine *renderer.Engine
}

func NewServer(engine *renderer.Engine) *Server {
	return &Server{engine: engine}
}
