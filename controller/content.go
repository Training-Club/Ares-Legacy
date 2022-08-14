package controller

import (
	"ares/database"
	"ares/model"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

// GetPostByID returns a single post object matching the provided ID
func (controller *AresController) GetPostByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		requestAccountId := ctx.GetString("accountId")

		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "id is not a hex"})
			return
		}

		post, err := database.FindDocumentById[model.Post](database.QueryParams{
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

		// TODO: Simplify this to not rely on string literal
		authorAccount, err := database.FindDocumentById[model.Account](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: "account",
		}, post.Author.Hex())

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "author account not found"})
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if authorAccount.Preferences.Privacy.ProfilePrivacy == model.PRIVATE {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if authorAccount.Preferences.Privacy.ProfilePrivacy == model.FOLLOWER_ONLY {
			reqAccountId, err := primitive.ObjectIDFromHex(requestAccountId)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "request account not found"})
					return
				}

				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to perform requesting account lookup: " + err.Error()})
				return
			}

			filter := bson.M{"followedId": authorAccount.ID, "followerId": reqAccountId}
			_, err = database.FindDocumentByFilter[model.Follow](database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: "follow",
			}, filter)

			if err != nil {
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		ctx.JSON(http.StatusOK, post)
	}
}

// GetPostsByQuery returns an array of posts matching the provided
// search query parameters
func (controller *AresController) GetPostsByQuery() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// GetCommentsByPostID returns a paginated list of comment objects
// matching the provided post ID
//
// 'key' param determines if we should look in the posts or
// the comments collection
func (controller *AresController) GetCommentsByPostID(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// GetCommentCount returns a comment count for a post
//
// 'key' param determines if we should look in the posts
// or the comments collection
func (controller *AresController) GetCommentCount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var postType model.PostItemType
		id := ctx.Param("id")

		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid post id hex"})
			return
		}

		postType = model.POST
		if key == "comment" {
			postType = model.COMMENT
		}

		count, err := database.Count(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{"post": id, "type": postType})

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": count})
	}
}

// GetLikeList returns a paginated list of like documents
// for  post matching the provided ID
//
// 'key' param determines if we should look in to the posts
// or the comments collection
func (controller *AresController) GetLikeList(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// GetLikeCount returns a count for comments on a provided post ID
//
// 'key' param determines if we should look in to the posts
// or the comments collection
func (controller *AresController) GetLikeCount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid post id hex"})
			return
		}

		count, err := database.Count(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{"post": id})

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": count})
	}
}

// CreatePost creates a new post object in the database
//
// If successful, a post ID will be returned with the document ID
// in a success 200 OK response
func (controller *AresController) CreatePost(s3Client *s3.Client, bucket string) gin.HandlerFunc {
	type Params struct {
		Location primitive.ObjectID  `json:"location,omitempty"`
		Text     string              `json:"text,omitempty"`
		Content  []model.ContentItem `json:"content" binding:"required"`
		Tags     []string            `json:"tags,omitempty"`
		Privacy  model.PrivacyLevel  `json:"privacy,omitempty"`
	}

	return func(ctx *gin.Context) {
		var params Params
		authorId := ctx.GetString("accountId")

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind params: " + err.Error()})
			return
		}

		authorHex, err := primitive.ObjectIDFromHex(authorId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid author id hex: " + err.Error()})
			return
		}

		if len(params.Content) > 10 {
			ctx.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}

		for _, content := range params.Content {
			exists, err := database.Exists(s3Client, bucket, content.Destination)

			if err != nil || !exists {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "file does not exist in datalake"})
				return
			}
		}

		post := model.Post{
			Author:    authorHex,
			Location:  params.Location,
			Text:      params.Text,
			Content:   params.Content,
			CreatedAt: time.Now(),
			Tags:      params.Tags,
			Privacy:   params.Privacy,
		}

		inserted, err := database.InsertOne[model.Post](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, post)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// CreateComment creates a new comment object in the database
//
// If successful, a comment ID will be returned with the document ID
// in a success 200 OK response
func (controller *AresController) CreateComment() gin.HandlerFunc {
	type Params struct {
		Post     primitive.ObjectID `json:"post" binding:"required"`
		PostType model.PostItemType `json:"type" binding:"required"`
		Text     string             `json:"text" binding:"required"`
	}

	return func(ctx *gin.Context) {
		var collectionName string
		var params Params

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind params: " + err.Error()})
			return
		}

		collectionName = "post"
		if params.PostType == model.COMMENT {
			collectionName = "comment"
		}

		authorId := ctx.GetString("accountId")
		authorIdHex, err := primitive.ObjectIDFromHex(authorId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "author id is not a valid hex"})
			return
		}

		// TODO: Check if authorId can make a comment on this post
		_, err = database.FindDocumentById(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: collectionName,
		}, params.Post.Hex())

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "post not found"})
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		comment := model.Comment{
			Post:      params.Post,
			Author:    authorIdHex,
			PostType:  params.PostType,
			Text:      params.Text,
			CreatedAt: time.Now(),
		}

		inserted, err := database.InsertOne[model.Comment](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, comment)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert comment document"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// AddLike creates a new like document in the database
//
// If successful, a like ID will be returned from the database
// in a success 200 OK response
func (controller *AresController) AddLike() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// RemoveLike deletes a like from the database
//
// Unlike other delete functions, this does not store the result in
// a 'deleted' version of the database as it is arbitrary to hold on to
func (controller *AresController) RemoveLike(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// UpdatePost performs an update on an existing Post document in the database
func (controller *AresController) UpdatePost() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// UpdateComment performs an update on an existing Comment document in the database
func (controller *AresController) UpdateComment() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// DeletePost queues a post to be deleted from the database
//
// If successful, the document will be removed from the database
// and moved to a 'deleted' version of the collection
// and a deleted ID will be returned in a success 200 OK response
func (controller *AresController) DeletePost() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// DeleteComment queues a comment to be deleted from the database
//
// If successful, the document will be removed from the database
// and moved to a 'deleted' version of the collection
// and a deleted ID will be returned in a success 200 OK response
func (controller *AresController) DeleteComment() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
