package routing

import (
	"ares/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyHealthCheckRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "",
		DatabaseName:   "",
	}

	router.GET("/status", ctrl.GetStatus())
}
