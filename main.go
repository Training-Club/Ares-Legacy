package main

import (
	"ares/config"
	"ares/database"
	"ares/routing"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	conf := config.Get()
	mongoClient, err := database.GetMongoClient(conf.Mongo.URI)
	if err != nil {
		panic("failed to establish mongo client instance: " + err.Error())
	}

	s3Client, err := database.GetS3Client(&database.S3Configuration{
		Key:      conf.S3.Key,
		Secret:   conf.S3.Secret,
		Token:    conf.S3.Token,
		Endpoint: conf.S3.Endpoint,
		Region:   conf.S3.Region,
	})

	redisClient, err := database.GetRedisClient(conf.Redis.Address, conf.Redis.Password, 0)
	if err != nil {
		panic("failed to establish redis cache instance: " + err.Error())
	}

	if err != nil {
		panic("failed to establish s3 client instance: " + err.Error())
		return
	}

	if conf.Gin.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// middleware
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// cors specific
	corsConfig := cors.DefaultConfig()

	if conf.Gin.Mode == "release" {
		corsConfig.AllowOrigins = []string{"https://www.trainingclubapp.com", "https://trainingclubapp.com", "http://localhost"}
	} else {
		corsConfig.AllowAllOrigins = true
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "HEAD"}
	corsConfig.AllowHeaders = []string{"Authorization", "Origin", "Token", "Content-Type"}
	corsConfig.ExposeHeaders = []string{"Set-Cookie", "Content-Length"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	routing.ApplyRoutes(router, mongoClient, s3Client, redisClient)

	err = router.Run(":" + conf.Gin.Port)
	if err != nil {
		panic("failed to start gin engine: " + err.Error())
	}
}
