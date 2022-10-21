package controller

import (
	"ares/database"
	"ares/model"
	"ares/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

// IsFollowing returns true if the provided followingId
// and followerId has an existing record in the database
func IsFollowing(
	mongoClient *mongo.Client,
	databaseName string,
	collectionName string,
	followingId primitive.ObjectID,
	followerId primitive.ObjectID,
) (bool, error) {
	filter := bson.M{
		"$and": bson.A{
			bson.M{"followingId": followingId},
			bson.M{"followedId": followerId},
		}}

	_, err := database.FindDocumentByFilter[model.Follow](database.QueryParams{
		MongoClient:    mongoClient,
		DatabaseName:   databaseName,
		CollectionName: collectionName,
	}, filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// IsFollowing returns a success 200 if the provided following id and
// followed account id have an existing follower record in the database
func (controller *AresController) IsFollowing() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		followingId := ctx.Param("followingId")
		followedId := ctx.Param("followedId")

		match := util.IsAlphanumeric(followingId)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "following id must be alphanumeric"})
			return
		}

		match = util.IsAlphanumeric(followedId)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "followed id must be alphanumeric"})
			return
		}

		followingHex, err := primitive.ObjectIDFromHex(followingId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "following id invalid"})
			return
		}

		followedHex, err := primitive.ObjectIDFromHex(followedId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "followed id invalid"})
			return
		}

		isFollowing, err := IsFollowing(controller.DB, controller.DatabaseName, controller.CollectionName, followingHex, followedHex)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query follow record: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": isFollowing})
	}
}

// GetConnectionCount can return the following/follower
// count for a provided account ID
func (controller *AresController) GetConnectionCount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if key != "followed" && key != "following" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key must be 'follower' or 'following'"})
			return
		}

		id := ctx.Param("id")
		match := util.IsAlphanumeric(id)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "id must be alphanumeric"})
			return
		}

		hex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "id is invalid hex"})
			return
		}

		count, err := database.Count(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{key + "Id": hex})

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// TODO: Check if requesting account id can see requested id's data

		ctx.JSON(http.StatusOK, gin.H{"result": count})
	}
}

func (controller *AresController) GetConnectionList(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if key != "followed" && key != "following" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key must be 'followed' or 'following'"})
			return
		}

		id := ctx.Param("id")
		page := ctx.DefaultQuery("page", "0")
		match := util.IsAlphanumeric(id)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "id must be alphanumeric"})
			return
		}

		hex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "id is invalid hex"})
			return
		}

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		result, err := database.FindManyDocumentsByFilterWithOpts[model.Follow](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{key + "Id": hex}, options.
			Find().
			SetLimit(100).
			SetSkip(int64(pageNumber*100)).
			SetSort(bson.D{{Key: "followedAt", Value: -1}}))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "no records"})
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": result})
	}
}

func (controller *AresController) GetMutualConnections(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func (controller *AresController) StartFollowing() gin.HandlerFunc {
	followDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	// TODO: Figure out a more elegant way to pass these credentials, as it
	// defeats the purpose of using the AresController for consistency
	accountDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "account",
	}

	return func(ctx *gin.Context) {
		followingId := ctx.GetString("accountId")
		followedId := ctx.Param("followedId")

		followingHex, err := primitive.ObjectIDFromHex(followingId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "following id is not a valid hex"})
			return
		}

		followedHex, err := primitive.ObjectIDFromHex(followedId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "followed id is not a valid hex"})
			return
		}

		if followingHex == followedHex {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "can not follow self"})
			return
		}

		filter := bson.M{"followingId": followingHex, "followedId": followedHex}
		existingRecord, err := database.FindDocumentByFilter[model.Follow](followDbQueryParams, filter)
		if err != nil && err != mongo.ErrNoDocuments {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !reflect.ValueOf(existingRecord).IsZero() {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "follow record already exists"})
			return
		}

		followedAccount, err := database.FindDocumentById[model.Account](accountDbQueryParams, followedId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "followed account not found"})
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var status model.FollowStatus
		if followedAccount.Preferences.Account.FollowRequestEnabled {
			status = model.PENDING
		} else {
			status = model.ACCEPTED
		}

		follow := model.Follow{
			FollowingID: followingHex,
			FollowedID:  followedHex,
			FollowedAt:  time.Now(),
			Status:      status,
		}

		inserted, err := database.InsertOne(followDbQueryParams, follow)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": inserted})
	}
}

func (controller *AresController) StopFollowing() gin.HandlerFunc {
	followDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		followingId := ctx.GetString("accountId")
		followedId := ctx.Param("followedId")

		_, err := primitive.ObjectIDFromHex(followingId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "following id is not a valid hex"})
			return
		}

		_, err = primitive.ObjectIDFromHex(followedId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "followed is is not a valid hex"})
			return
		}

		filter := bson.M{"$and": bson.A{bson.M{"followingId": followingId}, bson.M{"followedId": followedId}}}

		record, err := database.FindDocumentByFilter[model.Follow](followDbQueryParams, filter)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne(followDbQueryParams, record)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete record"})
			return
		}

		if deleteResult.DeletedCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "deleted count was less than 1"})
			return
		}

		ctx.Status(http.StatusOK)
	}
}
