package controller

import (
	"ares/database"
	"ares/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
	"time"
)

type MongoDatabaseInfo struct {
	DatabaseName   string
	CollectionName string
}

// query all posts from all accounts the provided account id follows
func getPostsFromFollowed(
	mongoClient *mongo.Client,
	followDatabase MongoDatabaseInfo,
	postDatabase MongoDatabaseInfo,
	accountId primitive.ObjectID,
	after time.Time,
	page uint8,
) ([]model.Post, error) {
	followedList, err := database.FindManyDocumentsByFilterWithOpts[model.Follow](database.QueryParams{
		MongoClient:    mongoClient,
		DatabaseName:   followDatabase.DatabaseName,
		CollectionName: followDatabase.CollectionName,
	}, bson.M{"followingId": accountId}, options.Find().SetLimit(10000))

	if err != nil {
		return nil, err
	}

	var followingIds []primitive.ObjectID
	for _, document := range followedList {
		followingIds = append(followingIds, document.FollowingID)
	}

	filter := bson.M{}
	filter["author"] = bson.M{"$in": followingIds}
	filter["createdAt"] = bson.M{"$gte": primitive.NewDateTimeFromTime(after)}

	posts, err := database.FindManyDocumentsByFilterWithOpts[model.Post](database.QueryParams{
		MongoClient:    mongoClient,
		DatabaseName:   postDatabase.DatabaseName,
		CollectionName: postDatabase.CollectionName,
	}, filter, options.Find().SetLimit(10).SetSkip(int64(page*10)).SetSort(bson.D{{Key: "createdAt", Value: -1}}))

	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetFeedContent queries posts to render in a users feed
func (controller *AresController) GetFeedContent() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid account id hex"})
			return
		}

		page := ctx.DefaultQuery("page", "0")
		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		defaultTime := time.Now()
		defaultTime = defaultTime.AddDate(0, 0, -3)
		afterTime := ctx.DefaultQuery("after", defaultTime.Format(time.RFC3339))
		after, err := time.Parse(time.RFC3339, afterTime)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid after timestamp"})
			return
		}

		postsFromFollowing, err := getPostsFromFollowed(
			controller.DB,
			MongoDatabaseInfo{
				DatabaseName:   "prod",
				CollectionName: "follow",
			},
			MongoDatabaseInfo{
				DatabaseName:   "prod",
				CollectionName: "post",
			},
			accountIdHex,
			after,
			uint8(pageNumber))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "no new posts"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query posts from followed accounts: " + err.Error()})
			return
		}

		// TODO: Remove this when we add additional queries
		if postsFromFollowing == nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "no posts found"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": postsFromFollowing})
	}
}
