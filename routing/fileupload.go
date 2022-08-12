package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyFileUploadRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "file",
		DatabaseName:   "prod",
	}

	v1Authorized := router.Group("/v1/fileupload")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.POST("/upload", ctrl.UploadFile())
	}
}
