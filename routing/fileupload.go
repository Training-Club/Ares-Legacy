package routing

import (
	"ares/config"
	"ares/controller"
	"ares/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyFileUploadRoutes(router *gin.Engine, mongoClient *mongo.Client, s3Client *s3.Client) {
	conf := config.Get()

	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "file",
		DatabaseName:   "prod",
	}

	v1Authorized := router.Group("/v1/fileupload")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.POST("/upload", ctrl.UploadFile(s3Client, conf.S3.Bucket))
	}
}
