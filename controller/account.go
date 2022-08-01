package controller

import (
	"ares/database"
	"ares/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

func getAccountWithKeyValue(
	controller *AresController,
	ctx *gin.Context,
	key string,
) (model.Account, error) {
	k := ctx.GetString(key)
	v := ctx.GetString("value")

	return database.FindDocumentByKeyValue[string, model.Account](database.QueryParams{
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
		MongoClient:    controller.DB,
	}, k, v)
}

func (controller *AresController) GetAccountAvailability() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.GetString("key")
		value := ctx.GetString("value")

		if key != "username" && key != "email" {
			ctx.AbortWithStatusJSON(400, gin.H{"message": "key field must be 'username' or 'email'"})
			return
		}

		_, err := database.FindDocumentByKeyValue[string, model.Account](
			database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: controller.CollectionName,
			}, key, value)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.Status(http.StatusOK)
				return
			}

			ctx.AbortWithStatus(http.StatusBadRequest)
		}

		ctx.Status(http.StatusConflict)
	}
}

func (controller *AresController) GetAccount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		account, err := getAccountWithKeyValue(controller, ctx, key)
		if err != nil {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"id": account.ID.Hex(), "username": account.Username})
	}
}

func (controller *AresController) GetProfile(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		account, err := getAccountWithKeyValue(controller, ctx, key)
		if err != nil || account.Preferences.Privacy.ProfilePrivacy == model.PRIVATE {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":       account.ID,
			"username": account.Username,
			"profile":  account.Profile,
		})
	}
}
