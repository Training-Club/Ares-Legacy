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
	"time"
)

// GetBlogById returns a single blog post matching the provided
// document ID
func (controller *AresController) GetBlogById() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "id is not a hex"})
			return
		}

		blog, err := database.FindDocumentById[model.BlogPost](database.QueryParams{
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

		ctx.JSON(http.StatusOK, blog)
	}
}

// GetBlogByQuery returns an array of blog posts matching the
// provided query
func (controller *AresController) GetBlogByQuery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		title, titlePresent := ctx.GetQuery("title")
		body, bodyPresent := ctx.GetQuery("body")
		tags, tagsPresent := ctx.GetQueryArray("tags")
		page := ctx.DefaultQuery("page", "0")

		if !titlePresent && !bodyPresent && !tagsPresent {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "need one of the required fields 'title', 'body' and 'tags'"})
			return
		}

		filter := bson.M{}

		if titlePresent {
			match := util.IsAlphanumeric(title)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "title text must be alphanumeric"})
				return
			}

			filter["title"] = primitive.Regex{Pattern: title, Options: "i"}
		}

		if bodyPresent {
			match := util.IsAlphanumeric(body)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "body text must be alphanumeric"})
				return
			}

			filter["body"] = primitive.Regex{Pattern: body, Options: "i"}
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

		result, err := database.FindManyDocumentsByFilterWithOpts[model.BlogPost](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter, options.Find().SetLimit(10).SetSkip(int64(pageNumber*10)).SetSort(bson.D{{Key: "createdAt", Value: -1}}))

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

// CreateBlog creates a new blog document in the database and
// returns the blog document ID in a response 200
func (controller *AresController) CreateBlog() gin.HandlerFunc {
	type Params struct {
		Title    string   `json:"title" binding:"required"`
		Subtitle string   `json:"subtitle,omitempty"`
		Body     string   `json:"body" binding:"required"`
		CoverUrl string   `json:"coverUrl,omitempty"`
		Tags     []string `json:"tags,omitempty"`
	}

	return func(ctx *gin.Context) {
		var params Params
		authorId := ctx.GetString("accountId")

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to bind params to struct"})
			return
		}

		authorHex, err := primitive.ObjectIDFromHex(authorId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid author id hex"})
			return
		}

		blog := model.BlogPost{
			Author:    authorHex,
			Title:     params.Title,
			Subtitle:  params.Subtitle,
			Body:      params.Body,
			CoverUrl:  params.CoverUrl,
			CreatedAt: time.Now(),
			Tags:      params.Tags,
		}

		inserted, err := database.InsertOne[model.BlogPost](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, blog)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// UpdateBlog updates an existing blog post with the
// newly provided schema
func (controller *AresController) UpdateBlog() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var blog model.BlogPost

		err := ctx.ShouldBindJSON(&blog)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal blog struct"})
			return
		}

		_, err = database.FindDocumentById[model.BlogPost](dbQueryParams, blog.ID.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		blog.EditedAt = time.Now()
		updatedCount, err := database.UpdateOne[model.BlogPost](dbQueryParams, blog.ID, blog)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update document: " + err.Error()})
			return
		}

		if updatedCount <= 0 {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Status(http.StatusOK)
	}
}

// DeleteBlog removes a blog document from the database
func (controller *AresController) DeleteBlog() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		blogId := ctx.Param("id")
		blogIdHex, err := primitive.ObjectIDFromHex(blogId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "blog id must be hex"})
			return
		}

		existingBlog, err := database.FindDocumentById[model.BlogPost](dbQueryParams, blogIdHex.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deletedBlog := model.DeletedBlogPost{
			Blog:      existingBlog,
			RemovalAt: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		}

		_, err = database.InsertOne[model.DeletedBlogPost](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, deletedBlog)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne[model.BlogPost](dbQueryParams, existingBlog)
		if deleteResult.DeletedCount <= 0 || err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.Status(http.StatusOK)
	}
}
