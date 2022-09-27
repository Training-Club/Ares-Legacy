package routing

import (
	"ares/controller"
	"ares/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyExerciseRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "exercise_sessions",
		DatabaseName:   "prod",
	}

	v1 := router.Group("/v1/exercise-session")
	{
		v1.GET("/count", ctrl.GetExerciseSessionCount())
	}

	v1Authorized := router.Group("/v1/exercise-session")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/id/:value", ctrl.GetExerciseSessionByID())
		v1Authorized.GET("/search", ctrl.GetExerciseSessionByQuery())

		v1Authorized.POST("/", ctrl.CreateExerciseSession())

		v1Authorized.PUT("/", ctrl.UpdateExerciseSession())

		v1Authorized.DELETE("/:sessionId", ctrl.DeleteExerciseSession())
	}
}
