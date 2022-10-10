package controller

import (
	"ares/audit"
	"ares/database"
	"ares/model"
	"ares/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
)

// GetRoles returns all roles currently in the database
func (controller *AresController) GetRoles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		attachedPermissions := ctx.Keys["attachedPermissions"].([]model.Permission)

		if !util.ContainsPerm(model.VIEW_ROLES, attachedPermissions) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "insufficient permissions"})
			return
		}

		roles, err := database.FindManyDocumentsByFilter[model.Role](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, bson.D{})

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query roles: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, roles)
	}
}

// GetRolesByAccount returns all roles attached to the provided account id
func (controller *AresController) GetRolesByAccount() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accountId := ctx.Param("accountId")
		_, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal account id"})
			return
		}

		account, err := database.FindDocumentById[model.Account](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, accountId)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "account not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query account: " + err.Error()})
			return
		}

		var roles []model.Role

		for _, roleId := range account.Roles {
			role, err := database.FindDocumentById[model.Role](database.QueryParams{
				MongoClient:    controller.DB,
				DatabaseName:   controller.DatabaseName,
				CollectionName: "role",
			}, roleId.Hex())

			if err != nil {
				continue
			}

			roles = append(roles, role)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": roles})
	}
}

// CreateRole reads CreateRoleParams and creates a new role document
// If successful the object id will be returned in a success 200 code
func (controller *AresController) CreateRole() gin.HandlerFunc {
	type CreateRoleParams struct {
		Name        string             `json:"name" binding:"required"`
		DisplayName string             `json:"displayName" binding:"required"`
		Permissions []model.Permission `json:"permissions,omitempty"`
	}

	roleDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal account id"})
			return
		}

		attachedPermissions := ctx.Keys["attachedPermissions"].([]model.Permission)

		if !util.ContainsPerm(model.GRANT_ROLES, attachedPermissions) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "insufficient permissions"})
			return
		}

		var params CreateRoleParams
		err = ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal request body"})
			return
		}

		match := util.IsAlphanumericWithWhitespace(params.Name)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "name must be alphanumeric"})
			return
		}

		match = util.IsAlphanumericWithWhitespace(params.DisplayName)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "display name must be alphanumeric"})
			return
		}

		_, err = database.FindDocumentByKeyValue[string, model.Role](roleDbQueryParams, "name", params.Name)

		if err != nil && err != mongo.ErrNoDocuments {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query existing role: " + err.Error()})
			return
		}

		role := model.Role{
			Name:        params.Name,
			DisplayName: params.DisplayName,
			Permissions: params.Permissions,
		}

		inserted, err := database.InsertOne[model.Role](roleDbQueryParams, role)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document: " + err.Error()})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.CREATE_ROLE,
			Context:     []string{"role name: " + role.Name},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// DeleteRole will attempt to remove a role from the database as well as remove the role
// from all users with it
func (controller *AresController) DeleteRole() gin.HandlerFunc {
	roleDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	accountDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "account",
	}

	return func(ctx *gin.Context) {
		attachedPermissions := ctx.Keys["attachedPermissions"].([]model.Permission)

		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal account id"})
			return
		}

		roleId := ctx.Param("roleId")
		_, err = primitive.ObjectIDFromHex(roleId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal role id"})
			return
		}

		if !util.ContainsPerm(model.GRANT_ROLES, attachedPermissions) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "insufficient permissions"})
			return
		}

		role, err := database.FindDocumentById[model.Role](roleDbQueryParams, roleId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "role not found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query existing role: " + err.Error()})
			return
		}

		deleteResult, err := database.DeleteOne[model.Role](roleDbQueryParams, role)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete role: " + err.Error()})
			return
		}

		if deleteResult.DeletedCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to delete role, delete count was zero"})
			return
		}

		accounts, err := database.FindManyDocumentsByFilter[model.Account](accountDbQueryParams, bson.M{"roles": bson.M{"$in": role.ID}})
		updateCount := 0

		for _, account := range accounts {
			var roleIds []primitive.ObjectID

			for _, roleId := range account.Roles {
				if roleId == role.ID {
					continue
				}

				roleIds = append(roleIds, roleId)
			}

			count, err := database.UpdateOne[model.Account](accountDbQueryParams, account.ID, account)
			if err != nil {
				continue
			}

			updateCount += int(count)
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.DELETE_ROLE,
			Context:     []string{"role name: " + role.Name, "affected accounts: " + strconv.FormatInt(int64(updateCount), 64)},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.JSON(http.StatusOK, gin.H{"result": updateCount})
	}
}

func (controller *AresController) GrantRole() gin.HandlerFunc {
	accountDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "account",
	}

	roleDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		attachedPermissions := ctx.Keys["attachedPermissions"].([]model.Permission)
		if !util.ContainsPerm(model.GRANT_ROLES, attachedPermissions) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "insufficient permissions"})
			return
		}

		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal account id"})
			return
		}

		grantedAccountId := ctx.Param("accountId")
		_, err = primitive.ObjectIDFromHex(grantedAccountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal granted account id"})
			return
		}

		grantedRoleId := ctx.Param("roleId")
		roleIdHex, err := primitive.ObjectIDFromHex(grantedRoleId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal granted role id"})
			return
		}

		_, err = database.FindDocumentById[model.Account](accountDbQueryParams, accountId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find account"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query account: " + err.Error()})
			return
		}

		grantedAccount, err := database.FindDocumentById[model.Account](accountDbQueryParams, grantedAccountId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find granted account"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query granted account: " + err.Error()})
			return
		}

		grantedRole, err := database.FindDocumentById[model.Role](roleDbQueryParams, grantedRoleId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find granted role"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query granted role: " + err.Error()})
			return
		}

		if util.Contains[primitive.ObjectID](roleIdHex, grantedAccount.Roles) {
			ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"message": "account has already been granted this role"})
			return
		}

		roleIds := grantedAccount.Roles
		roleIds = append(roleIds, roleIdHex)
		grantedAccount.Roles = roleIds

		updateCount, err := database.UpdateOne[model.Account](accountDbQueryParams, grantedAccount.ID, grantedAccount)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update account: " + err.Error()})
			return
		}

		if updateCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to update account: update count is zero"})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.GRANT_ROLE,
			Context:     []string{"role id: " + grantedRoleId, "role name: " + grantedRole.Name, "account id: " + grantedAccountId, "account username: " + grantedAccount.Username},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}

func (controller *AresController) RevokeRole() gin.HandlerFunc {
	accountDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: "account",
	}

	roleDbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		attachedPermissions := ctx.Keys["attachedPermissions"].([]model.Permission)
		if !util.ContainsPerm(model.GRANT_ROLES, attachedPermissions) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "insufficient permissions"})
			return
		}

		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal account id"})
			return
		}

		revokedAccountId := ctx.Param("accountId")
		_, err = primitive.ObjectIDFromHex(revokedAccountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal revoked account id"})
			return
		}

		revokedRoleId := ctx.Param("roleId")
		roleIdHex, err := primitive.ObjectIDFromHex(revokedRoleId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshal revoked role id"})
			return
		}

		_, err = database.FindDocumentById[model.Account](accountDbQueryParams, accountId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find account"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query account: " + err.Error()})
			return
		}

		revokedAccount, err := database.FindDocumentById[model.Account](accountDbQueryParams, revokedAccountId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find revoked account"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query revoked account: " + err.Error()})
			return
		}

		revokedRole, err := database.FindDocumentById[model.Role](roleDbQueryParams, revokedRoleId)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to find revoked role"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query revoked role: " + err.Error()})
			return
		}

		if !util.Contains[primitive.ObjectID](roleIdHex, revokedAccount.Roles) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "account does not have this role"})
			return
		}

		var roleIds []primitive.ObjectID

		for _, roleId := range revokedAccount.Roles {
			if roleId == revokedRole.ID {
				continue
			}

			roleIds = append(roleIds, roleId)
		}

		revokedAccount.Roles = roleIds

		updateCount, err := database.UpdateOne[model.Account](accountDbQueryParams, revokedAccount.ID, revokedAccount)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update account: " + err.Error()})
			return
		}

		if updateCount <= 0 {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "failed to update account: update count was zero"})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.REVOKE_ROLE,
			Context:     []string{"role id: " + revokedRoleId, "role name: " + revokedRole.Name, "account id: " + revokedAccountId, "account username: " + revokedAccount.Username},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}

func (controller *AresController) GrantRolePermission() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func (controller *AresController) RevokeRolePermission() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
