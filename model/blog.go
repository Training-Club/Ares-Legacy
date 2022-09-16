package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BlogPost struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Author    primitive.ObjectID `json:"author" bson:"author" binding:"required"`
	Slug      string             `json:"slug" bson:"slug" binding:"required"`
	Title     string             `json:"title" bson:"title" binding:"required"`
	Subtitle  string             `json:"subtitle,omitempty" bson:"subtitle,omitempty"`
	Body      string             `json:"body" bson:"body" binding:"required"`
	CoverUrl  string             `json:"coverUrl,omitempty" bson:"coverUrl,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt" binding:"required"`
	EditedAt  time.Time          `json:"editedAt,omitempty" bson:"editedAt,omitempty"`
	Tags      []string           `json:"tags,omitempty" bson:"tags,omitempty"`
}
