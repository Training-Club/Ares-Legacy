package controller

import (
	"ares/database"
	"ares/model"
	"ares/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"reflect"
)

func getAccountWithKeyValue(
	controller *AresController,
	ctx *gin.Context,
	key string,
) (model.Account, error) {
	v := ctx.Param("value")

	if key == "id" {
		id, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return model.Account{}, err
		}

		return database.FindDocumentByKeyValue[primitive.ObjectID, model.Account](database.QueryParams{
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
			MongoClient:    controller.DB,
		}, key, id)
	}

	return database.FindDocumentByKeyValue[string, model.Account](database.QueryParams{
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
		MongoClient:    controller.DB,
	}, key, v)
}

func (controller *AresController) GetAccountAvailability() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		value := ctx.Param("value")

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
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "account not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to find account: " + err.Error()})
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

// /v1/account/query/:username

func (controller *AresController) GetSimilarAccountsByUsername() gin.HandlerFunc {
	type BasicAccount struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	return func(ctx *gin.Context) {
		username := ctx.Param("username")

		filter := bson.M{"name": primitive.Regex{Pattern: username, Options: "i"}}
		accounts, err := database.FindManyDocumentsByFilter[model.Account](database.QueryParams{
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

		var basicAccounts []BasicAccount
		for _, account := range accounts {
			basic := BasicAccount{ID: account.ID.Hex(), Username: account.Username}
			basicAccounts = append(basicAccounts, basic)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": basicAccounts})
	}
}

func (controller *AresController) CreateStandardAccount() gin.HandlerFunc {
	type Params struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		CollectionName: controller.CollectionName,
		DatabaseName:   controller.DatabaseName,
	}

	return func(ctx *gin.Context) {
		var params Params

		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{
					"message": "failed to unmarshal json object: " + err.Error(),
				})

			return
		}

		existingEmail, _ := database.FindDocumentByKeyValue[string, model.Account](dbQueryParams, "email", params.Email)

		if !reflect.ValueOf(existingEmail).IsZero() {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "email is in use"})
			return
		}

		existingUsername, _ := database.FindDocumentByKeyValue[string, model.Account](dbQueryParams, "username", params.Username)

		if !reflect.ValueOf(existingUsername).IsZero() {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "username is in use"})
			return
		}

		var hashedPwd string
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(params.Password), 8)
		if hashErr != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to hash password: " + hashErr.Error()})
			return
		}

		hashedPwd = string(hash)

		acc := model.Account{
			Username: params.Username,
			Email:    params.Email,
			Password: hashedPwd,
		}

		id, err := database.InsertOne(dbQueryParams, acc)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document"})
			return
		}

		tokenString, err := util.GenerateToken(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to sign token"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":       id,
			"email":    acc.Email,
			"username": acc.Username,
			"type":     model.AccountType("standard"),
			"token":    tokenString,
		})
	}
}
