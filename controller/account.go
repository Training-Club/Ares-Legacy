package controller

import (
	"ares/audit"
	"ares/config"
	"ares/database"
	"ares/model"
	"ares/util"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Returns a single account matching key/value pair
// key = Document key
//
// value is derived from context params, searching for a :value string in the query
func getAccountWithKeyValue(
	controller *AresController,
	ctx *gin.Context,
	key string,
) (model.Account, error) {
	v := ctx.Param("value")
	match := util.IsAlphanumeric(v)

	if match {
		return model.Account{}, fmt.Errorf("value must be alphanumeric")
	}

	if key == "id" {
		return database.FindDocumentById[model.Account](database.QueryParams{
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
			MongoClient:    controller.DB,
		}, v)
	}

	return database.FindDocumentByKeyValue[string, model.Account](database.QueryParams{
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
		MongoClient:    controller.DB,
	}, key, v)
}

// Returns an array of accounts matching a
// similar string for the provided key/value pair
func getAccountsFuzzySearch(
	controller *AresController,
	key string,
	value string,
) ([]model.Account, error) {
	filter := bson.M{key: primitive.Regex{Pattern: value, Options: "i"}}
	accounts, err := database.FindManyDocumentsByFilter[model.Account](database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}, filter)

	return accounts, err
}

// GetAccountAvailability checks the database to see if the provided
// key/value pair is already in use in the database
//
// In the event an account already exists, the request will return a 409 Conflict
func (controller *AresController) GetAccountAvailability() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Param("key")
		match := util.IsAlphanumeric(key)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key must be alphanumeric"})
			return
		}

		value := ctx.Param("value")

		if key == "username" {
			match = util.IsAlphanumeric(value)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "value must be alphanumeric"})
				return
			}
		} else if key == "email" {
			match = util.IsValidEmail(value)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "email is invalid"})
				return
			}
		}

		if key != "username" && key != "email" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key field must be 'username' or 'email'"})
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

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to perform lookup: " + err.Error()})
			return
		}

		ctx.Status(http.StatusConflict)
	}
}

// GetAccount returns a single account matching the provided key/value
// pair provided through parameters and in the query itself.
//
// key = passed as a parameter, as this function is called from within the
//
//	router itself.
func (controller *AresController) GetAccount(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := ctx.Param("value")
		match := util.IsAlphanumeric(v)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "value must be alphanumeric"})
			return
		}

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

// GetProfile returns the basic account structure and profile struct
// attached to the provided key/value pair
func (controller *AresController) GetProfile(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestAccountId := ctx.GetString("accountId")
		account, err := getAccountWithKeyValue(controller, ctx, key)
		if err != nil {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		reqAccountId, err := primitive.ObjectIDFromHex(requestAccountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "request account id was invalid"})
			return
		}

		if reqAccountId != account.ID {
			// TODO: Perform blocked user check

			if account.Preferences.Privacy.ProfilePrivacy == model.FOLLOWER_ONLY {
				// TODO: Check if user follows another user
				// If the provided ids do not follow each other disallow
			} else if account.Preferences.Privacy.ProfilePrivacy == model.PRIVATE {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":       account.ID,
			"username": account.Username,
			"profile":  account.Profile,
		})
	}
}

// GetSimilarAccountsByUsername returns an array of accounts that match
// the provided username. This request utilizes a fuzzy search algo to query
// the results from the database
func (controller *AresController) GetSimilarAccountsByUsername() gin.HandlerFunc {
	type BasicAccount struct {
		ID       string        `json:"id"`
		Username string        `json:"username"`
		Profile  model.Profile `json:"profile"`
	}

	return func(ctx *gin.Context) {
		username := ctx.Param("username")
		accounts, err := getAccountsFuzzySearch(controller, "name", username)
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
			basic := BasicAccount{
				ID:       account.ID.Hex(),
				Username: account.Username,
				Profile:  account.Profile,
			}

			basicAccounts = append(basicAccounts, basic)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": basicAccounts})
	}
}

// GetSimilarAccountsByProfileName performs a fuzzy search to find all profiles
// that have profile names similar to the provided username.
//
// This function utilizes fuzzy search to return similar names
func (controller *AresController) GetSimilarAccountsByProfileName() gin.HandlerFunc {
	type BasicAccount struct {
		ID       string        `json:"id"`
		Username string        `json:"username"`
		Profile  model.Profile `json:"profile"`
	}

	return func(ctx *gin.Context) {
		name := ctx.Param("name")
		accounts, err := getAccountsFuzzySearch(controller, "profile.name", name)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to perform fuzzy search: " + err.Error()})
			return
		}

		var basicAccounts []BasicAccount
		for _, account := range accounts {
			basic := BasicAccount{
				ID:       account.ID.Hex(),
				Username: account.Username,
				Profile:  account.Profile,
			}

			basicAccounts = append(basicAccounts, basic)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": basicAccounts})
	}
}

// CreateStandardAccount creates a new account in the system
// using the 'standard' recipe.
//
// If successful an account ID will be generated by the database
// and returned to the request maker
func (controller *AresController) CreateStandardAccount(secure bool) gin.HandlerFunc {
	conf := config.Get()

	accessTokenPublicKey := conf.Auth.AccessTokenPublicKey
	accessTokenTTL := conf.Auth.AccessTokenTTL

	refreshTokenPublicKey := conf.Auth.RefreshTokenPublicKey
	refreshTokenTTL := conf.Auth.RefreshTokenTTL

	isReleaseVersion := conf.Gin.Mode == "release"

	type AccountResponse struct {
		ID       string            `json:"id"`
		Username string            `json:"username"`
		Email    string            `json:"email"`
		Type     model.AccountType `json:"type"`
	}

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

		sanitized := util.IsAlphanumeric(params.Username)
		if sanitized {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "username must be alphanumeric"})
			return
		}

		sanitized = util.IsValidEmail(params.Email)
		if sanitized {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "email is invalid"})
			return
		}

		validPassword := util.IsValidPassword(params.Password)
		if validPassword {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "password is invalid"})
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
			Type:     model.STANDARD,
		}

		id, err := database.InsertOne(dbQueryParams, acc)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document"})
			return
		}

		accessToken, err := util.GenerateToken(id, accessTokenPublicKey, accessTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate auth token"})
			return
		}

		refreshToken, err := util.GenerateToken(id, refreshTokenPublicKey, refreshTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to generate refresh token"})
			return
		}

		response := AccountResponse{
			ID:       id,
			Username: acc.Username,
			Email:    acc.Email,
			Type:     model.AccountType("standard"),
		}

		_, err = database.SetCacheValue(database.RedisClientParams{
			RedisClient: controller.RedisCache,
		}, refreshToken, id, refreshTokenTTL)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to cache refresh token"})
			return
		}

		idHex, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
				MongoClient: controller.DB,
				Initiator:   idHex,
				IP:          ctx.ClientIP(),
				EventName:   audit.CREATE_ACCOUNT,
				Context:     []string{"email: " + params.Email, "username: " + params.Username},
			})

			if err != nil {
				fmt.Println("failed to save audit entry: ", err)
			}
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

		ctx.JSON(http.StatusOK, gin.H{
			"account":       response,
			"token":         accessToken,
			"refresh_token": refreshToken,
		})
	}
}

// UpdateAccount updates struct data within the account struct such as
// profile, notifications, biometrics and privacy settings
//
// If successful the response will return an empty status OK 200
func (controller *AresController) UpdateAccount(id string) gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		if id != "notifications" &&
			id != "privacy" &&
			id != "profile" &&
			id != "biometrics" {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		accountId := ctx.GetString("accountId")
		account, err := database.FindDocumentById[model.Account](dbQueryParams, accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find account attached to request account id"})
			return
		}

		if id == "notifications" {
			var params model.NotificationPreferences

			err = ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal notification preferences object"})
				return
			}

			account.Preferences.Notifications = params
		} else if id == "privacy" {
			var params model.PrivacyPreferences

			err = ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal privacy preferences object"})
				return
			}

			account.Preferences.Privacy = params
		} else if id == "profile" {
			var params model.Profile

			err = ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal profile object"})
				return
			}

			account.Profile = params
		} else if id == "biometrics" {
			var params model.Biometrics

			err = ctx.ShouldBindJSON(&params)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal biometrics object"})
				return
			}

			account.Biometrics = params
		}

		updated, err := database.UpdateOne(dbQueryParams, account.ID, account)
		if updated <= 0 {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "updated document count returned as zero"})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   account.ID,
			IP:          ctx.ClientIP(),
			EventName:   audit.UPDATE_ACCOUNT,
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}

// SetAccountLastSeen sets the account last seen attached
// to the request headers to the current time on the server
func (controller *AresController) SetAccountLastSeen() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		account, err := database.FindDocumentById[model.Account](dbQueryParams, accountId)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		account.LastSeen = time.Now()

		updateCount, err := database.UpdateOne(dbQueryParams, account.ID, account)
		if err != nil || updateCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update document"})
			return
		}

		ctx.Status(http.StatusOK)
	}
}

// GetAccountCount returns an estimated count of documents in the
// accounts collection and returns it in a success 200
func (controller *AresController) GetAccountCount() gin.HandlerFunc {
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

// DeleteAccount will remove an account from the account database and
// transfer it to the deleted collection. Along with the account itself,
// a deleted account struct contains the time it will expire and need to
// be removed. That date is then picked up by a utility application which
// will clean up the database and handle the actual removal
func (controller *AresController) DeleteAccount() gin.HandlerFunc {
	accountDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		account, err := database.FindDocumentById[model.Account](accountDbQueryParams, accountId)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// TODO: Make removalAt customizable
		deletedAccount := model.DeletedAccount{
			Account:   account,
			RemovalAt: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		}

		deletedId, err := database.InsertOne(database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName + "_deleted",
		}, deletedAccount)

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne(accountDbQueryParams, account)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "account not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete account from database"})
			return
		}

		if deleteResult.DeletedCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "account delete result returned as zero"})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   account.ID,
			IP:          ctx.ClientIP(),
			EventName:   audit.DELETE_ACCOUNT,
			Context:     []string{"deleted id: " + deletedId},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.JSON(http.StatusOK, gin.H{"deletedId": deletedId})
	}
}
