package main

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/storage"
)

type ApiConfig struct {
	storageClient   *storage.Client
	firestoreClient *firestore.Client
	pubsubPublisher *pubsub.Publisher
	bucketName      string
}
