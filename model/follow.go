package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Follow struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FollowingID primitive.ObjectID `json:"followingId" bson:"followingId" binding:"required"`
	FollowedID  primitive.ObjectID `json:"followedId" bson:"followedId" binding:"required"`
	FollowedAt  time.Time          `json:"followedAt,omitempty" bson:"followedAt,omitempty"`
	Status      FollowStatus       `json:"status" bson:"status" binding:"required"`
}

type FollowStatus string

const (
	PENDING  FollowStatus = "PENDING"
	ACCEPTED FollowStatus = "ACCEPTED"
)
