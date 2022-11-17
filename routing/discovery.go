package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyDiscoveryRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	const DATABASE_NAME string = "prod"

	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "discovery",
		DatabaseName:   DATABASE_NAME,
	}

	v1Authorized := router.Group("/v1/discovery")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/query", ctrl.GetFeedContent())
	}
}
