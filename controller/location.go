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
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

// GetLocationById returns a single location document matching the
// provided document ID
func (controller *AresController) GetLocationById() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		_, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad location id hex"})
			return
		}

		location, err := database.FindDocumentById[model.Location](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, id)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, location)
	}
}

// GetLocationsByQuery returns an array of locations matching the
// provided query parameters including name, description similarities,
// location types and a radius from a specific lat/long
func (controller *AresController) GetLocationsByQuery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name, namePresent := ctx.GetQuery("name")
		description, descriptionPresent := ctx.GetQuery("description")
		types, typesPresent := ctx.GetQueryArray("type")
		lat, latPresent := ctx.GetQuery("lat")
		long, longPresent := ctx.GetQuery("long")
		maxDistance, maxDistancePresent := ctx.GetQuery("distance")
		page := ctx.DefaultQuery("page", "0")

		filter := bson.M{}

		if namePresent {
			match := util.IsAlphanumeric(name)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "name must be alphanumeric"})
				return
			}

			filter["name"] = primitive.Regex{Pattern: name, Options: "i"}
		}

		if descriptionPresent {
			match := util.IsAlphanumericWithWhitespace(description)
			if match {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "description must be alphanumeric"})
				return
			}

			filter["description"] = primitive.Regex{Pattern: description, Options: "i"}
		}

		if typesPresent {
			for _, t := range types {
				match := util.IsAlphanumeric(t)
				if match {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": t + " must be alphanumeric"})
					return
				}
			}

			filter["type"] = bson.M{"$in": types}
		}

		if latPresent && longPresent && maxDistancePresent {
			lat64, err := strconv.ParseFloat(lat, 64)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "latitude is not float64 format"})
				return
			}

			long64, err := strconv.ParseFloat(long, 64)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "longitude is not float64 format"})
				return
			}

			distance, err := strconv.Atoi(maxDistance)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "max distance field"})
			}

			filter["coordinates"] = bson.M{
				"$near": bson.M{
					"$geometry": model.Coordinates{
						Type:        "Point",
						Coordinates: []float64{long64, lat64},
					},
					"$maxDistance": distance,
				},
			}
		}

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		result, err := database.FindManyDocumentsByFilterWithOpts[model.Location](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter, options.Find().SetLimit(10).SetSkip(int64(pageNumber*10)))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// CreateLocation creates a new location in the database and returns
// the newly created documents Document ID
func (controller *AresController) CreateLocation() gin.HandlerFunc {
	type Params struct {
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description" binding:"required"`
		Type        model.LocationType `json:"type" binding:"required"`
		Coordinates model.Coordinates  `json:"coordinates,omitempty"`
	}

	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var params Params
		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshall params: " + err.Error()})
			return
		}

		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "account id hex invalid"})
			return
		}

		match := util.IsAlphanumericWithWhitespace(params.Name)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "location name must be alphanumeric"})
			return
		}

		match = util.IsAlphanumeric(params.Description)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "description must be alphanumeric"})
			return
		}

		if len(params.Coordinates.Coordinates) != 2 {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "coordinate array must be length of 2"})
			return
		}

		existing, err := database.FindDocumentByKeyValue[string, model.Location](dbQueryParams, "name", params.Name)

		if err != nil && err != mongo.ErrNoDocuments {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to look up existing document"})
			return
		}

		if !reflect.ValueOf(existing).IsZero() {
			ctx.AbortWithStatus(http.StatusConflict)
			return
		}

		location := model.Location{
			Author:      accountIdHex,
			Name:        params.Name,
			Description: params.Description,
			Type:        params.Type,
			Coordinates: params.Coordinates,
		}

		inserted, err := database.InsertOne[model.Location](dbQueryParams, location)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to insert document"})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   accountIdHex,
			IP:          ctx.ClientIP(),
			EventName:   audit.CREATE_LOCATION,
			Context:     []string{"location id: " + inserted, "location name: " + location.Name},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": inserted})
	}
}

// UpdateLocation accepts new parameters to update an existing
// document in the database
//
// If successful the response will return a 202 status accepted
func (controller *AresController) UpdateLocation() gin.HandlerFunc {
	type Params struct {
		ID          primitive.ObjectID `json:"id" binding:"required"`
		Name        string             `json:"name" binding:"required"`
		Description string             `json:"description" binding:"required"`
		Type        model.LocationType `json:"type" binding:"required"`
		Coordinates model.Coordinates  `json:"coordinates,omitempty"`
	}

	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		var params Params
		err := ctx.ShouldBindJSON(&params)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "failed to unmarshall params: " + err.Error()})
			return
		}

		accountId := ctx.GetString("accountId")
		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad account id hex"})
			return
		}

		match := util.IsAlphanumericWithWhitespace(params.Name)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "location name must be alphanumeric"})
			return
		}

		match = util.IsAlphanumericWithWhitespace(params.Description)
		if match {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "location description must be alphanumeric with spaces"})
			return
		}

		existing, err := database.FindDocumentById[model.Location](dbQueryParams, params.ID.Hex())

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !reflect.ValueOf(existing).IsZero() {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		if existing.Author != accountIdHex {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "must be location author to make edits"})
			return
		}

		existing.Name = params.Name
		existing.Description = params.Description
		existing.Type = params.Type
		existing.Coordinates = params.Coordinates

		inserted, err := database.UpdateOne[model.Location](dbQueryParams, existing.ID, existing)
		if err != nil || inserted <= 0 {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to update document"})
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   existing.Author,
			IP:          ctx.ClientIP(),
			EventName:   audit.UPDATE_LOCATION,
			Context:     []string{"location id: " + existing.ID.Hex(), "location name: " + existing.Name},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusAccepted)
	}
}

// DeleteLocation accepts a document ID and attempts
// to remove the document from the primary database
// then creates a Deleted Location entry in the
// deleted database
//
// If successful a success 200 response will be returned
func (controller *AresController) DeleteLocation() gin.HandlerFunc {
	dbQueryParams := database.QueryParams{
		MongoClient:    controller.DB,
		DatabaseName:   controller.DatabaseName,
		CollectionName: controller.CollectionName,
	}

	return func(ctx *gin.Context) {
		accountId := ctx.GetString("accountId")
		locationId := ctx.Param("id")

		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad account id hex"})
			return
		}

		locationIdHex, err := primitive.ObjectIDFromHex(locationId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad location id hex"})
			return
		}

		existing, err := database.FindDocumentById[model.Location](dbQueryParams, locationIdHex.Hex())
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if existing.Author != accountIdHex {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		deletedLocation := model.DeletedLocation{
			Location:  existing,
			RemovalAt: time.Now().Add(time.Hour * 24 * 7 * time.Duration(4)),
		}

		_, err = database.InsertOne[model.DeletedLocation](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName + "_deleted",
		}, deletedLocation)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		deleteResult, err := database.DeleteOne[model.Location](dbQueryParams, existing)
		if deleteResult.DeletedCount <= 0 || err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = audit.CreateAndSaveEntry(audit.CreateEntryParams{
			MongoClient: controller.DB,
			Initiator:   existing.Author,
			IP:          ctx.ClientIP(),
			EventName:   audit.DELETE_LOCATION,
			Context:     []string{"location id: " + existing.ID.Hex(), "location name: " + existing.Name},
		})

		if err != nil {
			fmt.Println("failed to save audit entry: ", err)
		}

		ctx.Status(http.StatusOK)
	}
}
