package controller

import "github.com/gin-gonic/gin"

// GetPostByID returns a single post object matching the provided ID
func (controller *AresController) GetPostByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {

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

	}
}

// CreatePost creates a new post object in the database
//
// If successful, a post ID will be returned with the document ID
// in a success 200 OK response
func (controller *AresController) CreatePost() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// CreateComment creates a new comment object in the database
//
// If successful, a comment ID will be returned with the document ID
// in a success 200 OK response
func (controller *AresController) CreateComment() gin.HandlerFunc {
	return func(ctx *gin.Context) {

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
