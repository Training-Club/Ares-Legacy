package routing

import (
	"ares/controller"
	"ares/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyExerciseInfoRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	const DATABASE_NAME string = "prod"

	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "exercise_info",
		DatabaseName:   "prod",
	}

	permissionHandler := middleware.PermissionMiddlewareHandler{
		MongoClient:           mongoClient,
		DatabaseName:          DATABASE_NAME,
		RoleCollectionName:    "role",
		AccountCollectionName: "account",
	}

	v1Authorized := router.Group("/v1/exercise-info")
	v1Authorized.Use(middleware.ValidateRequest(), permissionHandler.AttachPermissions())
	{
		v1Authorized.GET("/id/:value", ctrl.GetExerciseInfoByKeyValue("id"))
		v1Authorized.GET("/name/:value", ctrl.GetExerciseInfoByKeyValue("name"))
		v1Authorized.GET("/query", ctrl.QueryExerciseInfo())

		v1Authorized.POST("/", ctrl.CreateExerciseInfo())
	}
}
