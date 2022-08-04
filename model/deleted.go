package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeletedAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Account   Account            `json:"account,omitempty" bson:"account,omitempty" binding:"required"`
	RemovalAt time.Time          `json:"removalAt,omitempty" bson:"removalAt,omitempty" binding:"required"`
}
