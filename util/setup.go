package util

import (
	"ares/config"
	"ares/database"
	"ares/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// ConfigureAdminAccount reads a config value and attempts to create an initial
// admin account used to grant permissions to other users on the server
func ConfigureAdminAccount(mongoClient *mongo.Client, databaseName string, collectionName string) {
	conf := config.Get()
	shouldGenerateAdminAccount := conf.Ares.CreateAdminAccount

	accountDbQueryParams := database.QueryParams{
		MongoClient:    mongoClient,
		DatabaseName:   databaseName,
		CollectionName: collectionName,
	}

	if !shouldGenerateAdminAccount {
		account, err := database.FindDocumentByKeyValue[string, model.Account](accountDbQueryParams, "username", "admin")

		if err != nil && err != mongo.ErrNoDocuments {
			panic("failed to remove admin account: " + err.Error())
		}

		// catches mongo.ErrNoDocuments
		if err != nil {
			fmt.Println("failed to delete admin account: account not found")
			return
		}

		deleteResult, err := database.DeleteOne[model.Account](accountDbQueryParams, account)
		if err != nil {
			panic("failed to delete admin account: " + err.Error())
		}

		if deleteResult.DeletedCount <= 0 {
			panic("failed to delete admin account: delete result is zero")
		}

		fmt.Println("successfully deleted admin account")
		return
	}

	if gin.Mode() != gin.DebugMode {
		panic("attempted to create an admin account in production")
	}

	_, err := database.FindDocumentByKeyValue[string, model.Account](accountDbQueryParams, "username", "admin")
	if err != nil && err != mongo.ErrNoDocuments {
		panic("failed to create admin account: " + err.Error())
		return
	}

	if err != mongo.ErrNoDocuments {
		fmt.Println("failed to create admin account: account already exists")
		return
	}

	var hashedPwd string
	hash, hashErr := bcrypt.GenerateFromPassword([]byte("admin"), 8)
	if hashErr != nil {
		panic("failed to create admin account: " + hashErr.Error())
	}

	hashedPwd = string(hash)

	acc := model.Account{
		Username:    "admin",
		Email:       "admin@trainingclubapp.com",
		Password:    hashedPwd,
		Type:        model.STANDARD,
		Permissions: model.GetAllPermissions(),
	}

	_, err = database.InsertOne[model.Account](accountDbQueryParams, acc)
	if err != nil {
		panic("failed to create admin account: " + err.Error())
		return
	}

	fmt.Println("successfully created an admin account")
}
