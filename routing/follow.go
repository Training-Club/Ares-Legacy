package routing

import (
	"ares/controller"
	"ares/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

/*

	RETURN TRUE IF ACCOUNT ID FOLLOWING ACCOUNT ID DONE
	RETURN FOLLOWER COUNT FOR ACCOUNT ID DONE
	RETURN FOLLOWING COUNT FOR ACCOUNT ID DONE
	RETURN FOLLOWING ARRAY LIST (PAGINATED) FOR ACCOUNT ID DONE
	RETURN FOLLOWER ARRAY LIST (PAGINATED) FOR ACCOUNT ID DONE

*/

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
		v1Authorized.GET("/follower-count/:id", ctrl.GetConnectionCount("follower"))
		v1Authorized.GET("/following-count/:id", ctrl.GetConnectionCount("following"))
		v1Authorized.GET("/follower-list/:id", ctrl.GetConnectionList("follower"))
		v1Authorized.GET("/following-list/:id", ctrl.GetConnectionList("following"))
		v1Authorized.GET("/mutual/followers/:id", ctrl.GetMutualConnections("followers"))
		v1Authorized.GET("/mutual/following/:id", ctrl.GetMutualConnections("following"))

		v1Authorized.POST("/follow/:followedId", ctrl.StartFollowing())

		v1Authorized.DELETE("/unfollow/:followedId", ctrl.StopFollowing())
	}
}
