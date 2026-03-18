package main

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/storage"
	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

type WorkerConfig struct {
	engine           *renderer.Engine
	storageClient    *storage.Client
	firestoreClient  *firestore.Client
	pubsubSubscriber *pubsub.Subscriber
	bucketName       string
}
