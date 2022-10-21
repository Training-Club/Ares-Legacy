package controller

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type AresController struct {
	DB             *mongo.Client
	RedisCache     *redis.Client
	S3             *s3.Client
	DatabaseName   string
	CollectionName string
}
