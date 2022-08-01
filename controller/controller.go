package controller

import "go.mongodb.org/mongo-driver/mongo"

type AresController struct {
	DB             *mongo.Client
	DatabaseName   string
	CollectionName string
}
