package routing

import (
	"ares/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyBlogRoutes(router *gin.Engine, mongoClient *mongo.Client) {
	ctrl := controller.AresController{
		DB:             mongoClient,
		CollectionName: "blog",
		DatabaseName:   "prod",
	}

	v1 := router.Group("/v1/blog")
	{
		v1.GET("/:id", ctrl.GetBlogById())
		v1.GET("/query", ctrl.GetBlogByQuery())

		v1.POST("/", ctrl.CreateBlog())

		v1.PUT("/", ctrl.UpdateBlog())

		v1.DELETE("/", ctrl.DeleteBlog())
	}
}
