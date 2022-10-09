package controller

import (
	"ares/audit"
	"ares/database"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

func (controller *AresController) UploadFile(s3Client *s3.Client, bucket string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "failed to unmarshal account id"})
			return
		}

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

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.UPLOAD_FILE,
			Context:     ids,
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": ids})
	}
}
