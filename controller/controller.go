package controller

import (
	"github.com/go-redis/redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type AresController struct {
	DB             *mongo.Client
	RedisCache     *redis.Client
	DatabaseName   string
	CollectionName string
}
