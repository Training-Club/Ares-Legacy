package audit

import (
	"ares/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type CreateEntryParams struct {
	Initiator    primitive.ObjectID   `json:"initiator" binding:"required"`
	EventName    EntryType            `json:"eventName" binding:"required"`
	IP           string               `json:"ip,omitempty"`
	Context      []string             `json:"context,omitempty"`
	OtherParties []primitive.ObjectID `json:"otherParties,omitempty"`
	MongoClient  *mongo.Client
}

// CreateEntry accepts CreateEntryParams and returns an
// Entry struct populated with the params data
func CreateEntry(params CreateEntryParams) Entry {
	return Entry{
		Initiator:    params.Initiator,
		OtherParties: params.OtherParties,
		IP:           params.IP,
		EventName:    params.EventName,
		Timestamp:    time.Now(),
	}
}

// CreateAndSaveEntry creates a new entry using Create Entry Params
func CreateAndSaveEntry(params CreateEntryParams) error {
	entry := CreateEntry(params)

	_, err := database.InsertOne[Entry](database.QueryParams{
		MongoClient:    params.MongoClient,
		DatabaseName:   "prod",
		CollectionName: "audit",
	}, entry)

	return err
}
