package controller

import (
	"ares/database"
	"ares/model"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetExerciseInfoByKeyValue queries the database for a single key/value match and
// returns it in a success 200 if a match is found
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

		value := ctx.Param("value")

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

// QueryExerciseInfo accepts a query string of optional items and performs
// a lookup in the exercise info database for exercises that match the given
// criteria
//
// If no documents are found the response will be a 404, otherwise a successful 200 with
// an array of items under the "result" key
func (controller *AresController) QueryExerciseInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name, namePresent := ctx.GetQuery("name")
		muscleGroups, muscleGroupsPresent := ctx.GetQueryArray("muscleGroups")
		equipment, equipmentPresent := ctx.GetQuery("equipment")

		if !namePresent && !muscleGroupsPresent && !equipmentPresent {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		filter := bson.M{}

		if namePresent {
			filter["name"] = primitive.Regex{Pattern: name, Options: "i"}
		}

		if muscleGroupsPresent {
			filter["muscleGroups"] = bson.M{"$in": muscleGroups}
		}

		if equipmentPresent {
			filter["equipment"] = equipment
		}

		result, err := database.FindManyDocumentsByFilterWithOpts[model.ExerciseInfo](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter, options.Find().SetLimit(25).SetSort(bson.D{{Key: "verified", Value: -1}}))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to perform database query: " + err.Error()})
			return
		}

		if reflect.ValueOf(result).IsZero() {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": result})
	}
}

// CreateExerciseInfo reads parameters in and creates a new ExerciseInfo
// object within the database and if successful will return the inserted document
// in a success 200 response
func (controller *AresController) CreateExerciseInfo() gin.HandlerFunc {
	type Params struct {
		Name         string                  `json:"name" binding:"required"`
		Type         model.ExerciseType      `json:"type" binding:"required"`
		MuscleGroups []model.MuscleGroup     `json:"muscleGroups,omitempty"`
		Equipment    model.ExerciseEquipment `json:"equipment,omitempty"`
	}

	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var params Params

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal parameters"})
			return
		}

		existing, err := database.FindDocumentByKeyValue[string, model.Exercise](dbQueryParams, "name", params.Name)
		if err != nil && err != mongo.ErrNoDocuments {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !reflect.ValueOf(existing).IsZero() {
			ctx.AbortWithStatus(http.StatusConflict)
			return
		}

		data := model.ExerciseInfo{
			Name:         params.Name,
			Type:         params.Type,
			Verified:     false,
			MuscleGroups: params.MuscleGroups,
			Equipment:    params.Equipment,
		}

		inserted, err := database.InsertOne(dbQueryParams, data)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": inserted})
	}
}
