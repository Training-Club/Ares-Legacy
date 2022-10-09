package middleware

import (
	"ares/database"
	"ares/model"
	"ares/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type PermissionMiddlewareHandler struct {
	MongoClient           *mongo.Client
	DatabaseName          string
	AccountCollectionName string
	RoleCollectionName    string
}

// AttachPermissions reads a users permissions in to a simple
// array to make it is easier comparing permissions in handler functions
//
// Example on how to read the permissions back in from context:
// test := ctx.Keys["attachedPermissions"].([]model.Permission)
func (handler *PermissionMiddlewareHandler) AttachPermissions() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		_, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "bad authorization header"})
			return
		}

		account, err := database.FindDocumentById[model.Account](database.QueryParams{
			MongoClient:    handler.MongoClient,
			DatabaseName:   handler.DatabaseName,
			CollectionName: handler.AccountCollectionName,
		}, accountId)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to look up account during permission check"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "encountered an error while trying to look up account during permission check"})
			return
		}

		var permissions []model.Permission
		for _, permission := range account.Permissions {
			permissions = append(permissions, permission)
		}

		if len(account.Roles) > 0 {
			for _, roleId := range account.Roles {
				role, err := database.FindDocumentById[model.Role](database.QueryParams{
					MongoClient:    handler.MongoClient,
					DatabaseName:   handler.DatabaseName,
					CollectionName: handler.RoleCollectionName,
				}, roleId.Hex())

				if err != nil {
					continue
				}

				if len(role.Permissions) <= 0 {
					continue
				}

				for _, permission := range role.Permissions {
					if !util.ContainsPerm(permission, permissions) {
						permissions = append(permissions, permission)
					}
				}
			}
		}

		ctx.Set("attachedPermissions", permissions)
	}
}
