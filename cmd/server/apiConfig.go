package main

import (
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

type ApiConfig struct {
	engine        *renderer.Engine
	storageClient *storage.Client
	bucketName    string
}
