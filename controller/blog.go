package controller

import "github.com/gin-gonic/gin"

// GetBlogById returns a single blog post matching the provided
// document ID
func (controller *AresController) GetBlogById() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// GetBlogByQuery returns an array of blog posts matching the
// provided query
func (controller *AresController) GetBlogByQuery() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// CreateBlog creates a new blog document in the database and
// returns the blog document ID in a response 200
func (controller *AresController) CreateBlog() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// UpdateBlog updates an existing blog post with the
// newly provided schema
func (controller *AresController) UpdateBlog() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

// DeleteBlog removes a blog document from the database
func (controller *AresController) DeleteBlog() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
