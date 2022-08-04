package controller

import (
	"ares/database"
	"ares/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func (controller *AresController) GetExerciseInfoByKeyValue(key string) gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		CollectionName: controller.CollectionName,
		DatabaseName:   controller.DatabaseName,
	}

	return func(ctx *gin.Context) {
		if key != "id" && key != "name" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key must be 'id' or 'name'"})
			return
		}

		value := ctx.GetString("value")

		var info model.ExerciseInfo
		var err error

		if key == "id" {
			info, err = database.FindDocumentById[model.ExerciseInfo](dbQueryParams, value)
		} else if key == "name" {
			info, err = database.FindDocumentByKeyValue[string, model.ExerciseInfo](dbQueryParams, "name", value)
		}

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, info)
	}
}

// /v1/exerciseinfo/query/:name?muscleGroups=chest,tricep,bicep
func (controller *AresController) QueryExerciseInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func (controller *AresController) CreateExerciseInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
