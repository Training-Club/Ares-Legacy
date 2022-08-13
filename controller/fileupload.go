package controller

import (
	"ares/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (controller *AresController) UploadFile(s3Client *s3.Client, bucket string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		form, err := ctx.MultipartForm()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad multipart form data: " + err.Error()})
			return
		}

		files := form.File["upload[]"]
		ids := make([]string, len(files)/2)

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to read file header: " + err.Error()})
				return
			}

			fileSize := fileHeader.Size
			fileBuffer := make([]byte, fileSize)

			_, err = file.Read(fileBuffer)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to read file from buffer"})
				return
			}

			id, err := database.UploadFile(s3Client, bucket, fileBuffer)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to upload file to bucket: " + err.Error()})
				return
			}

			ids = append(ids, id)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": ids})
	}
}
