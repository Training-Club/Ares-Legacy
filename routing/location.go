package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyLocationRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "location",
		DatabaseName:   "prod",
	}

	v1Authorized := router.Group("/v1/location")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/id/:id", ctrl.GetLocationById())
		v1Authorized.GET("/search", ctrl.GetLocationsByQuery())

		v1Authorized.POST("/", ctrl.CreateLocation())

		v1Authorized.PUT("/", ctrl.UpdateLocation())

		v1Authorized.DELETE("/:id", ctrl.DeleteLocation())
	}
}
