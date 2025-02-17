package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyFollowRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "follow",
		DatabaseName:   "prod",
	}

	v1Authorized := router.Group("/v1/connections")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.GET("/is-following/:followingId/:followedId", ctrl.IsFollowing())
		v1Authorized.GET("/follower-count/:id", ctrl.GetConnectionCount("followed"))
		v1Authorized.GET("/following-count/:id", ctrl.GetConnectionCount("following"))
		v1Authorized.GET("/follower-list/:id", ctrl.GetConnectionList("followed"))
		v1Authorized.GET("/following-list/:id", ctrl.GetConnectionList("following"))
		v1Authorized.GET("/mutual/followers/:id", ctrl.GetMutualConnections("followed"))
		v1Authorized.GET("/mutual/following/:id", ctrl.GetMutualConnections("following"))

		v1Authorized.POST("/follow/:followedId", ctrl.StartFollowing())

		v1Authorized.DELETE("/unfollow/:followedId", ctrl.StopFollowing())
	}
}
