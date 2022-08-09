package database

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueryParams struct {
	MongoClient    *mongo.Client
	DatabaseName   string
	CollectionName string
}

func GetMongoClient(uri string) (*mongo.Client, error) {
	uri = strings.ReplaceAll(uri, "{CERT}", "certs/ca-certificate.crt")
	return mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
}

// FindDocumentById queries a single document in the database by document ID
func FindDocumentById[K any](params QueryParams, id string) (K, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var document K
	objectId, objectIdErr := primitive.ObjectIDFromHex(id)

	if objectIdErr != nil {
		return document, objectIdErr
	}

	findErr := collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&document)

	return document, findErr
}

// FindDocumentByKeyValue queries a single doucment in the database by a K/V filter
// Key must be a string, and value can be any filter object
func FindDocumentByKeyValue[K any, V any](params QueryParams, key string, value K) (V, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var document V
	findErr := collection.FindOne(ctx, bson.M{key: value}).Decode(&document)

	return document, findErr
}

// FindDocumentByFilter queries a single document in the daatabase using a BSON filter
func FindDocumentByFilter[K any](params QueryParams, filter bson.M) (K, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var document K
	findErr := collection.FindOne(ctx, filter).Decode(&document)

	return document, findErr
}

// FindManyDocumentsByKeyValue queries an array of documents by a K/V filter
func FindManyDocumentsByKeyValue[K any, V any](params QueryParams, key string, value K) ([]V, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var documents []V
	filterCursor, filterErr := collection.Find(ctx, bson.M{key: value})

	if filterErr != nil {
		return documents, filterErr
	}

	traverseErr := filterCursor.All(ctx, &documents)

	return documents, traverseErr
}

// FindManyDocumentsByFilter queries an array of documents by a BSON filter
func FindManyDocumentsByFilter[K any](params QueryParams, filter interface{}) ([]K, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var documents []K
	filterCursor, filterErr := collection.Find(ctx, filter)

	if filterErr != nil {
		return documents, filterErr
	}

	traverseErr := filterCursor.All(ctx, &documents)

	return documents, traverseErr
}

// FindManyDocumentsByFilterWithOpts queries an array of documents by a
// BSON filter with optional Find object options
func FindManyDocumentsByFilterWithOpts[K any](params QueryParams, filter interface{}, opts *options.FindOptions) ([]K, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var documents []K
	filterCursor, filterErr := collection.Find(ctx, filter, opts)

	if filterErr != nil {
		return documents, filterErr
	}

	traverseErr := filterCursor.All(ctx, &documents)

	return documents, traverseErr
}

// InsertOne adds a single document to the database
func InsertOne[K any](params QueryParams, document K) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	result, err := collection.InsertOne(ctx, document)

	if err != nil {
		return "", err
	}

	id := result.InsertedID.(primitive.ObjectID).Hex()

	return id, err
}

// UpdateOne updates a single document in the database
func UpdateOne[K any](params QueryParams, id primitive.ObjectID, document K) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": document})

	return result.ModifiedCount, err
}

// DeleteOne removes a single document from the database
func DeleteOne[K any](params QueryParams, document K) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	result, err := collection.DeleteOne(ctx, document)

	return result, err
}

// Count returns a number of documents matching the provided BSON filter
func Count(params QueryParams, filter interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	count, err := collection.CountDocuments(ctx, filter)

	return count, err
}
