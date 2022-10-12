package routing

import (
	"ares/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyPermissionRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	const DATABASE_NAME string = "prod"

	ctrl := controller.AresController{
		DB:             mongoClient,
		DatabaseName:   DATABASE_NAME,
		CollectionName: "permission",
	}

	v1 := router.Group("/v1/permission")
	{
		v1.GET("/account/:accountId", ctrl.GetPermissionListByAccount())
		v1.GET("/role/:roleId", ctrl.GetPermissionListByRole())
	}
}
