package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyRoleRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	const DATABASE_NAME string = "prod"

	ctrl := controller.AresController{
		DB:             mongoClient,
		DatabaseName:   DATABASE_NAME,
		CollectionName: "role",
	}

	permissionHandler := middleware.PermissionMiddlewareHandler{
		MongoClient:           mongoClient,
		DatabaseName:          DATABASE_NAME,
		RoleCollectionName:    "role",
		AccountCollectionName: "account",
	}

	v1Authorized := router.Group("/v1/role")
	v1Authorized.Use(middleware.ValidateRequest(), permissionHandler.AttachPermissions())
	{
		v1Authorized.GET("/", ctrl.GetRoles())
		v1Authorized.GET("/account/:accountId", ctrl.GetRolesByAccount())

		v1Authorized.POST("/", ctrl.CreateRole())

		v1Authorized.PUT("/grant/account/:accountId/:roleId", ctrl.GrantRole())
		v1Authorized.PUT("/grant/role/:roleId/:permissionName", ctrl.GrantRolePermission())
		v1Authorized.PUT("/revoke/role/:roleId/:permissionName", ctrl.RevokeRolePermission())

		v1Authorized.DELETE("/grant/account/:accountId/:roleId", ctrl.RevokeRole())
		v1Authorized.DELETE("/:roleId", ctrl.DeleteRole())
	}
}
