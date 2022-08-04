package routing

import (
	"ares/controller"
	"ares/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyExerciseInfoRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "exercise_info",
		DatabaseName:   "prod",
	}

	v1Authorized := router.Group("/v1/exercise-info")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/id/:value", ctrl.GetExerciseInfoByKeyValue("id"))
		v1Authorized.GET("/name/:value", ctrl.GetExerciseInfoByKeyValue("name"))
		v1Authorized.GET("/query/:value")

		v1Authorized.POST("/")
	}
}
