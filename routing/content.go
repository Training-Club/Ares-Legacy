package routing

import (
	"ares/config"
	"ares/controller"
	"ares/middleware"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyContentRoutes(router *gin.Engine, mongoClient *mongo.Client, s3Client *s3.Client) {
	conf := config.Get()

	postCtrl := controller.AresController{
		DB:             mongoClient,
		DatabaseName:   "prod",
		CollectionName: "post",
	}

	commentCtrl := controller.AresController{
		DB:             mongoClient,
		DatabaseName:   "prod",
		CollectionName: "comment",
	}

	likeCtrl := controller.AresController{
		DB:             mongoClient,
		DatabaseName:   "prod",
		CollectionName: "like",
	}

	v1Authorized := router.Group("/v1/content")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		// get post objects
		v1Authorized.GET("/post/id/:id", postCtrl.GetPostByID())
		v1Authorized.GET("/post/search", postCtrl.GetPostsByQuery())

		// get like list (paginated)
		v1Authorized.GET("/post/id/:id/likes", likeCtrl.GetLikeList("post"))
		v1Authorized.GET("/comment/id/:id/likes", likeCtrl.GetLikeList("comment"))

		// get like count
		v1Authorized.GET("/post/id/:id/likes/count", likeCtrl.GetLikeCount("post"))
		v1Authorized.GET("/comment/id/:id/likes/count", likeCtrl.GetLikeCount("comment"))

		// get comments (paginated)
		v1Authorized.GET("/post/id/:id/comments", commentCtrl.GetCommentsByPostID("post"))
		v1Authorized.GET("/comment/id/:id/comments", commentCtrl.GetCommentsByPostID("comment"))

		// create content, add likes
		v1Authorized.POST("/post/", postCtrl.CreatePost(s3Client, conf.S3.Bucket))
		v1Authorized.POST("/comment/", commentCtrl.CreateComment())
		v1Authorized.POST("/like/", likeCtrl.AddLike())

		// update content
		v1Authorized.PUT("/post/", postCtrl.UpdatePost())
		v1Authorized.PUT("/comment/", commentCtrl.UpdateComment())

		// delete content
		v1Authorized.DELETE("/post/:id", postCtrl.DeletePost())
		v1Authorized.DELETE("/comment/:id", commentCtrl.DeleteComment())

		// remove likes
		v1Authorized.DELETE("/like/post/:id", likeCtrl.RemoveLike("post"))
		v1Authorized.DELETE("/like/comment/:id", likeCtrl.RemoveLike("comment"))
	}
}
