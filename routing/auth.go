package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyAuthenticationRoutes(
	router *gin.Engine,
	mongoClient *mongo.Client,
	redisClient *redis.Client,
) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		RedisCache:     redisClient,
		CollectionName: "account",
		DatabaseName:   "prod",
	}

	v1 := router.Group("/v1/auth")
	{
		v1.GET("/:refreshToken", ctrl.RefreshToken())
		v1.POST("/", ctrl.AuthenticateStandardCredentials())
		v1.DELETE("/:refreshToken", ctrl.Logout())
	}

	v1Authorized := router.Group("/v1/auth")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/", ctrl.AuthenticateWithToken())
	}
}
