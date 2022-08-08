package controller

import (
	"ares/database"
	"ares/model"
	"ares/util"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetExerciseSessionByID retrieves a single exercise session document
// matching the provided doucment ID
func (controller *AresController) GetExerciseSessionByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("value")
		match := util.IsAlphanumeric(id)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "value must be alphanumeric"})
			return
		}

		session, err := database.FindDocumentById[model.Session](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, id)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, session)
	}
}

// GetExerciseSessionByQuery returns an array of sessions matching
// provided query strings
func (controller *AresController) GetExerciseSessionByQuery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionName, sessionNamePresent := ctx.GetQuery("name")
		before, beforePresent := ctx.GetQuery("before")
		exerciseName, exerciseNamePresent := ctx.GetQuery("exercise")
		page := ctx.DefaultQuery("page", "0")

		if !sessionNamePresent && !beforePresent && !exerciseNamePresent {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		filter := bson.M{}

		if sessionNamePresent {
			filter["sessionName"] = primitive.Regex{Pattern: sessionName, Options: "i"}
		}

		if beforePresent {
			beforeTime, err := time.Parse(before, "DD-MM-YYYY")
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid before time/date"})
				return
			}

			filter["createdAt"] = bson.M{"$gte": beforeTime}
		}

		if exerciseNamePresent {
			filter["exercises"] = bson.M{"$elemMatch": exerciseName}
		}

		pageNumber, err := strconv.ParseUint(page, 64, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad page number"})
			return
		}

		skip := int64(pageNumber * 25)

		result, err := database.FindManyDocumentsByFilterWithOpts[model.Exercise](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter, options.Find().SetLimit(25).SetSkip(skip))

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

func (controller *AresController) CreateExerciseSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func (controller *AresController) UpdateExerciseSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func (controller *AresController) DeleteExerciseSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
