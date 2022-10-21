package controller

import (
	"ares/audit"
	"ares/database"
	"ares/model"
	"ares/util"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
		authorId, authorIdPresent := ctx.GetQuery("author")
		text, textPresent := ctx.GetQuery("text")
		tags, tagsPresent := ctx.GetQueryArray("tags")
		page := ctx.DefaultQuery("page", "0")

		if !authorIdPresent && !textPresent && !tagsPresent {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "need one of the required fields: 'author', 'text' or 'tags"})
			return
		}

		filter := bson.M{}

		if authorIdPresent {
			authorIdHex, err := primitive.ObjectIDFromHex(authorId)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "author id invalid hex"})
				return
			}

			filter["author"] = authorIdHex
		}

		if textPresent {
			match := util.IsAlphanumeric(text)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "search text body must be alphanumeric"})
				return
			}

			filter["text"] = primitive.Regex{Pattern: text, Options: "i"}
		}

		if tagsPresent {
			for _, tag := range tags {
				match := util.IsAlphanumeric(tag)
				if match {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "tags must be alphanumeric"})
					return
				}
			}

			filter["tags"] = bson.M{"$in": tags}
		}

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		result, err := database.FindManyDocumentsByFilterWithOpts[model.Post](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter, options.
			Find().
			SetLimit(10).
			SetSkip(int64(pageNumber*10)).
			SetSort(bson.D{{Key: "createdAt", Value: -1}}))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": result})
	}
}

// GetCommentsByPostID returns a paginated list of comment objects
// matching the provided post ID
func (controller *AresController) GetCommentsByPostID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		postId := ctx.Param("id")
		page := ctx.DefaultQuery("page", "0")

		postIdHex, err := primitive.ObjectIDFromHex(postId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "post id invalid hex"})
			return
		}

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		comments, err := database.FindManyDocumentsByFilterWithOpts[model.Comment](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{"post": postIdHex}, options.Find().SetLimit(10).SetSkip(int64(pageNumber*10)).SetSort(bson.D{{Key: "createdAt", Value: -1}}))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "comments not found"})
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": comments})
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

		idHex, err := primitive.ObjectIDFromHex(id)
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
		}, bson.M{"post": idHex, "type": postType})

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
		if key != "post" && key != "comment" {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		postId := ctx.Param("id")
		page := ctx.DefaultQuery("page", "0")
		postIdHex, err := primitive.ObjectIDFromHex(postId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad post id hex"})
			return
		}

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		likes, err := database.FindManyDocumentsByFilterWithOpts[model.Like](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{"post": postIdHex}, options.Find().SetLimit(50).SetSkip(int64(pageNumber*50)))

		ctx.JSON(http.StatusOK, gin.H{"result": likes})
	}
}

// GetLikeCount returns a count for comments on a provided post ID
//
// 'key' param determines if we should look in to the posts
// or the comments collection
func (controller *AresController) GetLikeCount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		idHex, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid post id hex"})
			return
		}

		count, err := database.Count(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{"post": idHex})

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
		Session  primitive.ObjectID  `json:"session,omitempty"`
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
			Session:   params.Session,
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

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   authorHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.CREATE_POST,
			Context:     []string{"post id: " + inserted},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
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
		var params Params

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind params: " + err.Error()})
			return
		}

		authorId := ctx.GetString("accountId")
		authorIdHex, err := primitive.ObjectIDFromHex(authorId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "author id is not a valid hex"})
			return
		}

		// TODO: Check if authorId can make a comment on this post
		if params.PostType == model.POST {
			_, err = database.FindDocumentById[model.Post](database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: "post",
			}, params.Post.Hex())
		} else if params.PostType == model.COMMENT {
			_, err = database.FindDocumentById[model.Comment](database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: "comment",
			}, params.Post.Hex())
		}

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

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   authorIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.CREATE_COMMENT,
			Context:     []string{"comment id: " + inserted, "content: " + params.Text},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// AddLike creates a new like document in the database
//
// If successful, a like ID will be returned from the database
// in a success 200 OK response
func (controller *AresController) AddLike() gin.HandlerFunc {
	type Params struct {
		Post     primitive.ObjectID `json:"post" binding:"required"`
		PostType model.PostItemType `json:"type" binding:"required"`
	}

	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var params Params

		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad account id hex"})
			return
		}

		err = ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal params: " + err.Error()})
			return
		}

		// TODO: Check if user can see this post
		if params.PostType == model.POST {
			_, err := database.FindDocumentById[model.Post](database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: "post",
			}, params.Post.Hex())

			if err != nil {
				if err == mongo.ErrNoDocuments {
					ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "post not found"})
					return
				}

				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		} else if params.PostType == model.COMMENT {
			_, err := database.FindDocumentById[model.Comment](database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: "comment",
			}, params.Post.Hex())

			if err != nil {
				if err == mongo.ErrNoDocuments {
					ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "comment not found"})
					return
				}

				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		filter := bson.M{"post": params.Post, "author": accountIdHex}
		existingRecord, err := database.FindDocumentByFilter[model.Like](dbQueryParams, filter)

		if err != nil && err != mongo.ErrNoDocuments {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to check for existing like record: " + err.Error()})
			return
		}

		if !reflect.ValueOf(existingRecord).IsZero() {
			ctx.AbortWithStatus(http.StatusConflict)
			return
		}

		like := model.Like{
			Author:   accountIdHex,
			Post:     params.Post,
			PostType: params.PostType,
			LikedAt:  time.Now(),
		}

		inserted, err := database.InsertOne(dbQueryParams, like)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// RemoveLike deletes a like from the database
//
// Unlike other delete functions, this does not store the result in
// a 'deleted' version of the database as it is arbitrary to hold on to
func (controller *AresController) RemoveLike() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		postId := ctx.Param("id")

		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad account id hex"})
			return
		}

		postIdHex, err := primitive.ObjectIDFromHex(postId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad post id hex"})
		}

		filter := bson.M{"post": postIdHex, "author": accountIdHex}
		existingLike, err := database.FindDocumentByFilter[model.Like](dbQueryParams, filter)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deletedResult, err := database.DeleteOne(dbQueryParams, existingLike)
		if err != nil || deletedResult.DeletedCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to delete document"})
			return
		}

		ctx.Status(http.StatusOK)
	}
}

// UpdatePost performs an update on an existing Post document in the database
func (controller *AresController) UpdatePost() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var post model.Post

		err := ctx.ShouldBindJSON(&post)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal post object: " + err.Error()})
			return
		}

		_, err = database.FindDocumentById[model.Post](dbQueryParams, post.ID.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		post.EditedAt = time.Now()
		updatedCount, err := database.UpdateOne[model.Post](dbQueryParams, post.ID, post)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update document: "})
			return
		}

		if updatedCount <= 0 {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   post.Author,
			IP:          ctx.ClientIP(),
			EventName:   audit.UPDATE_POST,
			Context:     []string{"post id: " + post.ID.Hex()},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}

// UpdateComment performs an update on an existing Comment document in the database
func (controller *AresController) UpdateComment() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var comment model.Comment

		err := ctx.ShouldBind(&comment)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal comment struct: " + err.Error()})
			return
		}

		_, err = database.FindDocumentById[model.Comment](dbQueryParams, comment.ID.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		comment.EditedAt = time.Now()
		updatedCount, err := database.UpdateOne(dbQueryParams, comment.ID, comment)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update comment document: " + err.Error()})
			return
		}

		if updatedCount <= 0 {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   comment.Author,
			IP:          ctx.ClientIP(),
			EventName:   audit.UPDATE_COMMENT,
			Context:     []string{"comment id: " + comment.ID.Hex() + " content: " + comment.Text},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}

// GetPostCount returns an estimated count of documents in the
// post collection and returns it in a success 200
func (controller *AresController) GetPostCount() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		count, err := database.Count(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.M{})

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"result": 0})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"result": count})
	}
}

// DeletePost queues a post to be deleted from the database
//
// If successful, the document will be removed from the database
// and moved to a 'deleted' version of the collection
// and a deleted ID will be returned in a success 200 OK response
func (controller *AresController) DeletePost() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		postId := ctx.Param("id")

		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "account id is invalid hex"})
			return
		}

		postIdHex, err := primitive.ObjectIDFromHex(postId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "post id is invalid hex"})
			return
		}

		existingPost, err := database.FindDocumentById[model.Post](dbQueryParams, postIdHex.Hex())

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if existingPost.Author != accountIdHex {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		deletedPost := model.DeletedPost{
			Post:      existingPost,
			RemovalAt: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		}

		_, err = database.InsertOne[model.DeletedPost](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName + "_deleted",
		}, deletedPost)

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne[model.Post](dbQueryParams, existingPost)
		if deleteResult.DeletedCount <= 0 || err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.DELETE_POST,
			Context:     []string{"post id: " + existingPost.ID.Hex()},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}

// DeleteComment queues a comment to be deleted from the database
//
// If successful, the document will be removed from the database
// and moved to a 'deleted' version of the collection
// and a deleted ID will be returned in a success 200 OK response
func (controller *AresController) DeleteComment() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		commentId := ctx.Param("id")

		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "account id is invalid hex"})
			return
		}

		commentIdHex, err := primitive.ObjectIDFromHex(commentId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "comment id is invalid hex"})
			return
		}

		existingComment, err := database.FindDocumentById[model.Comment](dbQueryParams, commentIdHex.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if existingComment.Author != accountIdHex {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		deletedComment := model.DeletedComment{
			Comment:   existingComment,
			RemovalAt: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		}

		_, err = database.InsertOne[model.DeletedComment](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName + "_deleted",
		}, deletedComment)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne[model.Comment](dbQueryParams, existingComment)
		if deleteResult.DeletedCount <= 0 || err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.DELETE_COMMENT,
			Context:     []string{"comment id: " + existingComment.ID.Hex()},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}
