package routing

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyRoutes(engine *gin.Engine, mongoClient *mongo.Client) {
	ApplyAccountRoutes(engine, mongoClient)
	ApplyExerciseInfoRoutes(engine, mongoClient)
}
