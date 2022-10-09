package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Role struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name" binding:"required"`
	DisplayName string             `json:"displayName" bson:"displayName" binding:"required"`
	Permissions []Permission       `json:"permissions,omitempty" bson:"permissions,omitempty"`
}
