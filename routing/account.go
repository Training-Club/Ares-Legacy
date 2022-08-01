package routing

import (
	"ares/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyAccountRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "account",
		DatabaseName:   "prod",
	}

	v1 := router.Group("/v1/account")
	{
		v1.GET("/id/:value", ctrl.GetAccount("id"))
		v1.GET("/username/:value", ctrl.GetAccount("username"))
		v1.GET("/availability/:key/:value", ctrl.GetAccountAvailability())
		v1.GET("/profile/id/:value", ctrl.GetProfile("id"))
		v1.GET("/profile/username/:value", ctrl.GetProfile("username"))

		v1.POST("/recipe/standard")
		v1.POST("/recipe/apple")
		v1.POST("/recipe/google")
	}

	v1Authorized := router.Group("/v1/account")
	{
		v1Authorized.GET("/")
		v1Authorized.PUT("/")
		v1Authorized.DELETE("/")
	}
}
