package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Account struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username    string             `json:"username,omitempty" bson:"username,omitempty" binding:"required"`
	Email       string             `json:"email,omitempty" bson:"email,omitempty" binding:"required"`
	Password    string             `json:"password,omitempty" bson:"password,omitempty"`
	CreatedAt   time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty" binding:"required"`
	LastSeen    time.Time          `json:"lastSeen,omitempty" bson:"lastSeen,omitempty" binding:"required"`
	Type        AccountType        `json:"accountType,omitempty" bson:"accountType,omitempty" binding:"required"`
	Profile     Profile            `json:"profile,omitempty" bson:"profile,omitempty"`
	Biometrics  Biometrics         `json:"biometrics,omitempty" bson:"biometrics,omitempty"`
	Preferences Preferences        `json:"preferences,omitempty" bson:"preferences,omitempty"`
}

type Profile struct {
	Avatar   string `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Name     string `json:"name,omitempty" bson:"name,omitempty"`
	Location string `json:"location,omitempty" bson:"location,omitempty"`
	Bio      string `json:"bio,omitempty" bson:"bio,omitempty"`
}

type Biometrics struct {
	Birthday time.Time `json:"birthday,omitempty" bson:"birthday,omitempty"`
	Sex      string    `json:"sex,omitempty" bson:"sex,omitempty"`
	Weight   float32   `json:"weight,omitempty" bson:"weight,omitempty"`
	Height   float32   `json:"height,omitempty" bson:"height,omitempty"`
}

type NotificationPreferences struct {
	NotifyNewFollower        bool `json:"notifyNewFollower,omitempty" bson:"notifyNewFollower,omitempty"`
	NotifyNewLike            bool `json:"notifyNewLike,omitempty" bson:"notifyNewLike,omitempty"`
	NotifyNewMessage         bool `json:"notifyNewMessage,omitempty" bson:"notifyNewMessage,omitempty"`
	NotifyNewAssignedSession bool `json:"notifyNewAssignedSession,omitempty" bson:"notifyNewAssignedSession"`
	NotifyNewAssignedMeal    bool `json:"notifyNewAssignedMeal,omitempty" bson:"notifyNewAssignedMeal,omitempty"`
}

type PrivacyPreferences struct {
	ProfilePrivacy PrivacyLevel `json:"profilePrivacy,omitempty" bson:"profilePrivacy,omitempty"`
	MessagePrivacy PrivacyLevel `json:"messagePrivacy,omitempty" bson:"messagePrivacy,omitempty"`
	CommentPrivacy PrivacyLevel `json:"commentPrivacy,omitempty" bson:"commentPrivacy,omitempty"`
}

type AccountPreferences struct {
	FollowRequestEnabled bool `json:"followRequestEnabled,omitempty" bson:"followRequestEnabled,omitempty"`
}

type Preferences struct {
	Account       AccountPreferences      `json:"accountPreferences,omitempty" bson:"accountPreferences,omitempty"`
	Privacy       PrivacyPreferences      `json:"privacyPreferences,omitempty" bson:"privacyPreferences,omitempty"`
	Notifications NotificationPreferences `json:"notificationPreferences,omitempty" bson:"notificationPreferences,omitempty"`
}

type PrivacyLevel string

const (
	PUBLIC        PrivacyLevel = "public"
	FOLLOWER_ONLY PrivacyLevel = "follower_only"
	PRIVATE       PrivacyLevel = "private"
)

type AccountType string

const (
	STANDARD AccountType = "standard"
	APPLE    AccountType = "apple"
	GOOGLE   AccountType = "google"
)
