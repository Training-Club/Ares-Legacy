package routing

import (
	"ares/controller"
	"ares/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

/*

	GET
	- get by id
	- get by similar name
	- get by query
		- author id
		- before (date)
		- after (date)
		- contains exercise by name


	POST
	- create session

	PUT
	- update session

	DELETE
	- queue session to be deleted

*/

func ApplyExerciseRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "exercise_sessions",
		DatabaseName:   "prod",
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
