package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeletedAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Account   Account            `json:"account" bson:"account" binding:"required"`
	RemovalAt time.Time          `json:"removalAt" bson:"removalAt" binding:"required"`
}

type DeletedSession struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Session   Session            `json:"session" bson:"session" binding:"required"`
	RemovalAt time.Time          `json:"removalAt" bson:"removalAt" binding:"required"`
}

type DeletedPost struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Post      Post               `json:"post" bson:"post" binding:"required"`
	RemovalAt time.Time          `json:"removalAt" bson:"removalAt" binding:"required"`
}

type DeletedComment struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Comment   Comment            `json:"comment" bson:"comment" binding:"required"`
	RemovalAt time.Time          `json:"removalAt" bson:"removalAt" binding:"required"`
}

type DeletedLocation struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Location  Location           `json:"location" bson:"location" binding:"required"`
	RemovalAt time.Time          `json:"removalAt" bson:"removalAt" binding:"required"`
}
