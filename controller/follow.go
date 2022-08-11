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
	"strconv"
)

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

		filter := bson.M{
			"$and": bson.A{
				bson.M{"followingId": followingHex},
				bson.M{"followedId": followedHex},
			}}

		record, err := database.FindDocumentByFilter[model.Follow](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": record})
	}
}

func (controller *AresController) GetConnectionCount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if key != "follower" && key != "following" {
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

		count, err := database.Count[model.Follow](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{key: hex})

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
		if key != "follower" && key != "following" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key must be 'follower' or 'following'"})
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
		}, bson.M{key: hex}, options.
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
	return func(ctx *gin.Context) {

	}
}

func (controller *AresController) StopFollowing() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
