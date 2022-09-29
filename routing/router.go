package routing

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyRoutes(
	engine *gin.Engine,
	mongoClient *mongo.Client,
	s3Client *s3.Client,
	redisClient *redis.Client,
) {
	ApplyHealthCheckRoutes(engine, mongoClient)
	ApplyAuthenticationRoutes(engine, mongoClient, redisClient)
	ApplyAccountRoutes(engine, mongoClient)
	ApplyExerciseInfoRoutes(engine, mongoClient)
	ApplyExerciseRoutes(engine, mongoClient)
	ApplyFollowRoutes(engine, mongoClient)
	ApplyContentRoutes(engine, mongoClient, s3Client)
	ApplyLocationRoutes(engine, mongoClient)
	ApplyFileUploadRoutes(engine, mongoClient, s3Client)
	ApplyBlogRoutes(engine, mongoClient)
}
