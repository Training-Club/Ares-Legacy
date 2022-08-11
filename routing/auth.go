package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyAuthenticationRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "account",
		DatabaseName:   "prod",
	}

	v1 := router.Group("/v1/auth")
	{
		v1.POST("/", ctrl.AuthenticateStandardCredentials())
	}

	v1Authorized := router.Group("/v1/auth")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/", ctrl.AuthenticateWithToken())
	}
}
