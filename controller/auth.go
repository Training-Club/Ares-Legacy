package controller

import (
	"ares/database"
	"ares/model"
	"ares/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// AuthenticateWithToken authenticates a token attached to the
// current request headers and returns a status OK with basic
// account information
func (controller *AresController) AuthenticateWithToken() gin.HandlerFunc {
	type BasicAccount struct {
		ID       string            `json:"id"`
		Username string            `json:"username"`
		Email    string            `json:"email"`
		Type     model.AccountType `json:"type"`
	}

	return func(ctx *gin.Context) {
		id := ctx.GetString("accountId")
		account, err := database.FindDocumentById[model.Account](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, id)

		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		basic := BasicAccount{
			ID:       account.ID.Hex(),
			Username: account.Username,
			Email:    account.Email,
			Type:     account.Type,
		}

		ctx.JSON(http.StatusOK, basic)
	}
}

// AuthenticateStandardCredentials authenticates an email/password
// and generates a new JWT if the password matches
func (controller *AresController) AuthenticateStandardCredentials() gin.HandlerFunc {
	type BasicAccount struct {
		ID       string            `json:"id"`
		Username string            `json:"username"`
		Email    string            `json:"email"`
		Type     model.AccountType `json:"type"`
	}

	type Params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(ctx *gin.Context) {
		var params Params
		err := ctx.ShouldBindJSON(&params)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal credentials object"})
			return
		}

		account, err := database.FindDocumentByKeyValue[string, model.Account](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, "email", params.Email)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(params.Password))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "password does not match"})
			return
		}

		tokenString, err := util.GenerateToken(account.ID.Hex())
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate auth token"})
			return
		}

		basic := BasicAccount{
			ID:       account.ID.Hex(),
			Username: account.Username,
			Email:    account.Email,
			Type:     account.Type,
		}

		ctx.JSON(http.StatusOK, gin.H{"account": basic, "token": tokenString})
	}
}
