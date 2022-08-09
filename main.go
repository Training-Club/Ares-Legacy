package main

import (
	"ares/config"
	"ares/database"
	"ares/routing"
	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.Get()
	mongoClient, err := database.GetMongoClient(conf.Mongo.URI)
	if err != nil {
		panic("failed to establish mongo client instance: " + err.Error())
	}

	if conf.Gin.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	routing.ApplyRoutes(router, mongoClient)

	err = router.Run(":8080")
	if err != nil {
		panic("failed to start gin engine: " + err.Error())
	}
}
