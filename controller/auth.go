package controller

import (
	"ares/audit"
	"ares/config"
	"ares/database"
	"ares/middleware"
	"ares/model"
	"ares/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func (controller *AresController) AuthenticateStandardCredentials(secure bool) gin.HandlerFunc {
	conf := config.Get()

	accessTokenPublicKey := conf.Auth.AccessTokenPublicKey
	accessTokenTTL := conf.Auth.AccessTokenTTL

	refreshTokenPublicKey := conf.Auth.RefreshTokenPublicKey
	refreshTokenTTL := conf.Auth.RefreshTokenTTL

	isReleaseVersion := conf.Gin.Mode == "release"

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

		accessToken, err := util.GenerateToken(account.ID.Hex(), accessTokenPublicKey, accessTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate auth token"})
			return
		}

		refreshToken, err := util.GenerateToken(account.ID.Hex(), refreshTokenPublicKey, refreshTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate refresh token"})
			return
		}

		basic := BasicAccount{
			ID:       account.ID.Hex(),
			Username: account.Username,
			Email:    account.Email,
			Type:     account.Type,
		}

		_, err = database.SetCacheValue(database.RedisClientParams{
			RedisClient: controller.RedisCache,
		}, refreshToken, account.ID.Hex(), refreshTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to cache refresh token: " + err.Error()})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   account.ID,
			IP:          ctx.ClientIP(),
			EventName:   audit.AUTH_WITH_CREDENTIALS,
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		if secure {
			var cookieDomain string
			if isReleaseVersion {
				cookieDomain = "*.trainingclubapp.com"
			} else {
				cookieDomain = ".localhost"
			}

			ctx.SetSameSite(http.SameSiteNoneMode)
			ctx.SetCookie(
				"refresh_token",
				refreshToken,
				refreshTokenTTL,
				"/",
				cookieDomain,
				true,
				true,
			)
		}

		ctx.JSON(http.StatusOK, gin.H{"account": basic, "token": accessToken, "refresh_token": refreshToken})
	}
}

// RefreshToken takes an existing refresh_token from the query params
// and performs the following comparisons:
//   - Verify that the token is a valid JWT
//   - Query Redis Cache by Refresh Token for accountId value
//   - Verify that the accountId belongs to an existing account
//   - Generates a new access_token and returns in a success 200 response
func (controller *AresController) RefreshToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conf := config.Get()
		accessTokenPublicKey := conf.Auth.AccessTokenPublicKey
		accessTokenTTL := conf.Auth.AccessTokenTTL
		refreshPublicKey := conf.Auth.RefreshTokenPublicKey

		refreshToken, err := ctx.Cookie("refresh_token")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to read refresh_token cookie"})
			return
		}

		_, err = middleware.ValidateToken(refreshToken, refreshPublicKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal refresh token"})
			return
		}

		accountId, err := database.GetCacheValue(database.RedisClientParams{
			RedisClient: controller.RedisCache,
		}, refreshToken)
		if err != nil || accountId == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "failed to verify refresh token integrity"})
			return
		}

		_, err = database.FindDocumentById[model.Account](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, accountId)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find account"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to fin account"})
			return
		}

		newAccessToken, err := util.GenerateToken(accountId, accessTokenPublicKey, accessTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate new access token"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
	}
}

// Logout accepts a refresh_token then invalidates it in the cache
// then returns a success 200
func (controller *AresController) Logout(secure bool) gin.HandlerFunc {
	conf := config.Get()
	refreshPublicKey := conf.Auth.RefreshTokenPublicKey
	isReleaseVersion := conf.Gin.Mode == "release"

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		refreshToken, err := ctx.Cookie("refresh_token")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to read refresh_token cookie"})
			return
		}

		_, err = middleware.ValidateToken(refreshToken, refreshPublicKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal refresh token"})
			return
		}

		deleteCount, err := database.DeleteCacheValue(database.RedisClientParams{
			RedisClient: controller.RedisCache,
		}, refreshToken)

		if err != nil || deleteCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to delete from cache"})
			return
		}

		var cookieDomain string
		if isReleaseVersion {
			cookieDomain = "*.trainingclubapp.com"
		} else {
			cookieDomain = ".localhost"
		}

		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err == nil {
			err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
				MongoClient: controller.DB,
				Initiator:   accountIdHex,
				IP:          ctx.ClientIP(),
				EventName:   audit.LOGOUT,
			})

			if err != nil {
				fmt.Println("failed to save audit entry: ", err)
			}
		}

		if secure {
			ctx.SetCookie(
				"refresh_token",
				refreshToken,
				-1,
				"/",
				cookieDomain,
				true,
				true,
			)
		}

		ctx.Status(http.StatusOK)
	}
}
