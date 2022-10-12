package controller

import (
	"ares/database"
	"ares/model"
	"ares/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// GetPermissionListByAccount returns a list of all permissions assigned to the
// provided account id
//
// /v1/permission/account/:accountId
func (controller *AresController) GetPermissionListByAccount() gin.HandlerFunc {
	accountDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "account",
	}

	roleDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "role",
	}

	return func(ctx *gin.Context) {
		accountId := ctx.Param("accountId")
		_, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "failed to unmarshal account id"})
			return
		}

		account, err := database.FindDocumentById[model.Account](accountDbQueryParams, accountId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "account not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query account: " + err.Error()})
			return
		}

		var permissions []model.Permission

		for _, permission := range account.Permissions {
			permissions = append(permissions, permission)
		}

		for _, roleId := range account.Roles {
			role, err := database.FindDocumentById[model.Role](roleDbQueryParams, roleId.Hex())

			if err != nil {
				continue
			}

			for _, permission := range role.Permissions {
				if util.ContainsPerm(permission, permissions) {
					continue
				}

				permissions = append(permissions, permission)
			}
		}

		ctx.JSON(http.StatusOK, permissions)
	}
}

// GetPermissionListByRole returns a list of all permissions assigned to the
// provided role id
//
// /v1/permission/role/:roleId
func (controller *AresController) GetPermissionListByRole() gin.HandlerFunc {
	roleDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "role",
	}

	return func(ctx *gin.Context) {
		roleId := ctx.Param("roleId")
		_, err := primitive.ObjectIDFromHex(roleId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "failed to unmarshal role id"})
			return
		}

		role, err := database.FindDocumentById[model.Role](roleDbQueryParams, roleId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "role not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query role: " + err.Error()})
			return
		}

		var permissions []model.Permission
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission)
		}

		ctx.JSON(http.StatusOK, permissions)
	}
}
