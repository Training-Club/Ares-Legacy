package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Post struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Author    primitive.ObjectID `json:"author" bson:"author" binding:"required"`
	Location  primitive.ObjectID `json:"location,omitempty" bson:"location,omitempty"`
	Session   primitive.ObjectID `json:"session,omitempty" bson:"session,omitempty"`
	Text      string             `json:"text,omitempty" bson:"text,omitempty"`
	Content   []ContentItem      `json:"content" bson:"content" binding:"required"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt" binding:"required"`
	EditedAt  time.Time          `json:"editedAt,omitempty" bson:"editedAt,omitempty"`
	Tags      []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	Privacy   PrivacyLevel       `json:"privacy,omitempty" bson:"privacy,omitempty"`
}

type ContentItem struct {
	Destination string      `json:"destination" bson:"destination" binding:"required"`
	Type        ContentType `json:"type" bson:"type" binding:"required"`
}

type Like struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Author   primitive.ObjectID `json:"author" bson:"author" binding:"required"`
	Post     primitive.ObjectID `json:"post" bson:"post" binding:"required"`
	PostType PostItemType       `json:"type" bson:"type" binding:"required"`
	LikedAt  time.Time          `json:"likedAt" bson:"likedAt" binding:"required"`
}

type Comment struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Post      primitive.ObjectID `json:"post" bson:"post" binding:"required"`
	Author    primitive.ObjectID `json:"author" bson:"author" binding:"required"`
	PostType  PostItemType       `json:"type" bson:"type" binding:"required"`
	Text      string             `json:"text" bson:"text" binding:"required"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt" binding:"required"`
	EditedAt  time.Time          `json:"editedAt,omitempty" bson:"editedAt,omitempty"`
}

type ContentType string
type PostItemType string

const (
	IMAGE ContentType = "IMAGE"
	VIDEO ContentType = "VIDEO"
)

const (
	POST    PostItemType = "POST"
	COMMENT PostItemType = "COMMENT"
)
