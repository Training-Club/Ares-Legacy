package routing

import (
	"ares/controller"
	"ares/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyAccountRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	const DATABASE_NAME string = "prod"

	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "account",
		DatabaseName:   DATABASE_NAME,
	}

	permissionHandler := middleware.PermissionMiddlewareHandler{
		MongoClient:           mongoClient,
		DatabaseName:          DATABASE_NAME,
		RoleCollectionName:    "role",
		AccountCollectionName: "account",
	}

	v1 := router.Group("/v1/account")
	{
		v1.GET("/availability/:key/:value", ctrl.GetAccountAvailability())
		v1.GET("/count", ctrl.GetAccountCount())

		v1.POST("/recipe/standard", ctrl.CreateStandardAccount(false))
		v1.POST("/recipe/standard/secure", ctrl.CreateStandardAccount(true))

		v1.POST("/recipe/apple")
		v1.POST("/recipe/google")
	}

	v1Authorized := router.Group("/v1/account")
	v1Authorized.Use(middleware.ValidateRequest(), permissionHandler.AttachPermissions())
	{
		v1Authorized.GET("/username/:value", ctrl.GetAccount("username"))
		v1Authorized.GET("/id/:value", ctrl.GetAccount("id"))
		v1Authorized.GET("/search/:username", ctrl.GetSimilarAccountsByUsername())
		v1Authorized.GET("/profile/id/:value", ctrl.GetProfile("id"))
		v1Authorized.GET("/profile/username/:value", ctrl.GetProfile("username"))

		v1Authorized.PUT("/lastseen", ctrl.SetAccountLastSeen())
		v1Authorized.PUT("/preferences/notifications", ctrl.UpdateAccount("notifications"))
		v1Authorized.PUT("/preferences/privacy", ctrl.UpdateAccount("privacy"))
		v1Authorized.PUT("/preferences/profile", ctrl.UpdateAccount("profile"))
		v1Authorized.PUT("/preferences/biometrics", ctrl.UpdateAccount("biometrics"))

		v1Authorized.DELETE("/", ctrl.DeleteAccount())
	}
}
