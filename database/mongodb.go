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

func FindDocumentByKeyValue[K any, V any](params QueryParams, key string, value K) (V, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var document V
	findErr := collection.FindOne(ctx, bson.M{key: value}).Decode(&document)

	return document, findErr
}

func FindDocumentByFilter[K any](params QueryParams, filter bson.M) (K, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	var document K
	findErr := collection.FindOne(ctx, filter).Decode(&document)

	return document, findErr
}

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

func UpdateOne[K any](params QueryParams, id primitive.ObjectID, document K) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": document})

	return result.ModifiedCount, err
}

func DeleteOne[K any](params QueryParams, document K) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	result, err := collection.DeleteOne(ctx, document)

	return result, err
}

func Count(params QueryParams, filter interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	collection := params.MongoClient.Database(params.DatabaseName).Collection(params.CollectionName)

	defer cancel()

	count, err := collection.CountDocuments(ctx, filter)

	return count, err
}
