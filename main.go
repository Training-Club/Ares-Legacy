package ares

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

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	routing.ApplyRoutes(router, mongoClient)
}
