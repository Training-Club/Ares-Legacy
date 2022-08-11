package routing

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyRoutes(engine *gin.Engine, mongoClient *mongo.Client) {
	ApplyHealthCheckRoutes(engine, mongoClient)
	ApplyAuthenticationRoutes(engine, mongoClient)
	ApplyAccountRoutes(engine, mongoClient)
	ApplyExerciseInfoRoutes(engine, mongoClient)
	ApplyExerciseRoutes(engine, mongoClient)
}
