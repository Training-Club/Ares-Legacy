package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Follow struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	FollowerID primitive.ObjectID `json:"followerId" bson:"followerId" binding:"required"`
	FollowedID primitive.ObjectID `json:"followedId" bson:"followedId" binding:"required"`
	FollowedAt time.Time          `json:"followedAt,omitempty" bson:"followedAt,omitempty"`
	Status     FollowStatus       `json:"status" bson:"status" binding:"required"`
}

type FollowStatus string

const (
	PENDING  FollowStatus = "PENDING"
	ACCEPTED FollowStatus = "ACCEPTED"
)
