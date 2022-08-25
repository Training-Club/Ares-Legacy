package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Location struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Author      primitive.ObjectID `json:"author" bson:"author" binding:"required"`
	Name        string             `json:"name" bson:"name" binding:"required"`
	Description string             `json:"description" bson:"description" binding:"required"`
	Type        LocationType       `json:"type,omitempty" bson:"type,omitempty"`
	Coordinates Coordinates        `json:"coordinates,omitempty" bson:"coordinates,omitempty"`
}

type Coordinates struct {
	Type        string    `json:"type" bson:"type" binding:"required"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates" binding:"required"`
}

type LocationType string

const (
	GYM   string = "GYM"
	CITY  string = "CITY"
	PLACE string = "PLACE"
)
