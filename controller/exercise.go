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
// matching the provided document ID
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
		exerciseNames, exerciseNamesPresent := ctx.GetQueryArray("exercise")
		page := ctx.DefaultQuery("page", "0")

		if !sessionNamePresent && !beforePresent && !exerciseNamesPresent {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		filter := bson.M{}

		if sessionNamePresent {
			match := util.IsAlphanumeric(sessionName)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "session name must be alphanumeric"})
				return
			}

			filter["sessionName"] = primitive.Regex{Pattern: sessionName, Options: "i"}
		}

		if beforePresent {
			beforeTime, err := time.Parse("01-02-2006", before)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid before time/date: " + err.Error()})
				return
			}

			filter["createdAt"] = bson.M{"$gte": beforeTime}
		}

		if exerciseNamesPresent {
			for _, exerciseName := range exerciseNames {
				match := util.IsAlphanumeric(exerciseName)
				if match {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "exercise name must be alphanumeric"})
					return
				}
			}

			filter["exercises.exerciseName"] = bson.M{"$in": exerciseNames}
		}

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to retrieve page number: " + err.Error()})
			return
		}

		skip := int64(pageNumber * 25)

		result, err := database.FindManyDocumentsByFilterWithOpts[model.Session](database.QueryParams{
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

// CreateExerciseSession marshals request params and attempts to create a new
// training session in the database.
//
// If successful, resulting session document ID will be returned in a
// status 200 OK response
func (controller *AresController) CreateExerciseSession() gin.HandlerFunc {
	type Params struct {
		SessionName string              `json:"sessionName" binding:"required"`
		Author      primitive.ObjectID  `json:"author" binding:"required"`
		Status      model.SessionStatus `json:"status" binding:"required"`
		Timestamp   time.Time           `json:"timestamp,omitempty"`
		Exercises   []model.Exercise    `json:"exercises,omitempty" binding:"required"`
	}

	return func(ctx *gin.Context) {
		var params Params
		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal params: " + err.Error()})
			return
		}

		session := model.Session{
			SessionName: params.SessionName,
			Author:      params.Author,
			Status:      params.Status,
			Timestamp:   params.Timestamp,
			Exercises:   params.Exercises,
		}

		inserted, err := database.InsertOne(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, session)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": inserted})
	}
}

func (controller *AresController) UpdateExerciseSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// DeleteExerciseSession accepts a session id as a query param and
// attempts to queue the document for deletion in the database
//
// If successful, the document will be moved to a _deleted version of
// the collection and return a status 200 OK with the ID of the deleted document
func (controller *AresController) DeleteExerciseSession() gin.HandlerFunc {
	trainingDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		sessionId := ctx.Param("sessionId")

		match := util.IsAlphanumeric(sessionId)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "session id must be alphanumeric"})
			return
		}

		session, err := database.FindDocumentById[model.Session](trainingDbQueryParams, sessionId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if session.Author.Hex() != accountId {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// TODO: Make removalAt customizable
		deletedSession := model.DeletedSession{
			Session:   session,
			RemovalAt: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		}

		deletedId, err := database.InsertOne(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName + "_deleted",
		}, deletedSession)

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne(trainingDbQueryParams, session)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "session not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete document: " + err.Error()})
			return
		}

		if deleteResult.DeletedCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "session delete count is zero"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"deletedId": deletedId})
	}
}
