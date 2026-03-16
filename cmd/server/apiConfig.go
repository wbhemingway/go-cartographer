package main

import (
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
	"cloud.google.com/go/firestore"
)

type ApiConfig struct {
	engine        *renderer.Engine
	storageClient *storage.Client
	firestoreClient *firestore.Client
	bucketName    string
}
