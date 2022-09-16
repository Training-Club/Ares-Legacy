package routing

import (
	"ares/controller"
	"ares/middleware"
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
		v1.GET("/id/:id", ctrl.GetBlogById())
		v1.GET("/query", ctrl.GetBlogByQuery())
		v1.GET("/slug/:slug", ctrl.GetBlogBySlug())
	}

	v1Authorized := router.Group("/v1/blog")
	v1Authorized.Use(middleware.ValidateRequest())
	{
		v1Authorized.POST("/", ctrl.CreateBlog())

		v1Authorized.PUT("/", ctrl.UpdateBlog())

		v1Authorized.DELETE("/:id", ctrl.DeleteBlog())
	}
}
