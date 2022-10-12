package audit

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Entry struct {
	ID           primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Initiator    primitive.ObjectID   `json:"initiator" bson:"initiator" binding:"required"`
	IP           string               `json:"ip,omitempty" bson:"ip,omitempty"`
	OtherParties []primitive.ObjectID `json:"otherParties,omitempty" bson:"otherParties,omitempty"`
	Context      []string             `json:"context,omitempty" bson:"context,omitempty"`
	EventName    EntryType            `json:"eventName" bson:"eventName" binding:"required"`
	Timestamp    time.Time            `json:"timestamp" bson:"timestamp" binding:"required"`
}

type EntryType string

const (
	CREATE_ACCOUNT            EntryType = "create_account"
	UPDATE_ACCOUNT            EntryType = "update_account"
	DELETE_ACCOUNT            EntryType = "delete_account"
	AUTH_WITH_CREDENTIALS     EntryType = "auth_with_credentials"
	LOGOUT                    EntryType = "logout"
	CREATE_POST               EntryType = "create_post"
	CREATE_COMMENT            EntryType = "create_comment"
	UPDATE_POST               EntryType = "update_post"
	UPDATE_COMMENT            EntryType = "update_comment"
	DELETE_POST               EntryType = "delete_post"
	DELETE_COMMENT            EntryType = "delete_comment"
	CREATE_TRAINING_SESSION   EntryType = "create_training_session"
	UPDATE_TRAINING_SESSION   EntryType = "update_training_session"
	DELETE_TRAINING_SESSION   EntryType = "delete_training_session"
	CREATE_EXERCISE           EntryType = "create_exercise"
	UPLOAD_FILE               EntryType = "upload_file"
	CREATE_LOCATION           EntryType = "create_location"
	UPDATE_LOCATION           EntryType = "update_location"
	DELETE_LOCATION           EntryType = "delete_location"
	CREATE_ROLE               EntryType = "create_role"
	DELETE_ROLE               EntryType = "delete_role"
	UPDATE_ROLE               EntryType = "update_role"
	GRANT_ROLE                EntryType = "grant_role"
	REVOKE_ROLE               EntryType = "revoke_role"
	GRANT_ROLE_PERMISSION     EntryType = "grant_role_permission"
	REVOKE_ROLE_PERMISSION    EntryType = "revoke_role_permission"
	GRANT_ACCOUNT_PERMISSION  EntryType = "grant_account_permission"
	REVOKE_ACCOUNT_PERMISSION EntryType = "revoke_account_permission"
)
