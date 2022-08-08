package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type DietEntry struct {
	ID      primitive.ObjectID
	AddedAt time.Time
	Food    Food
}

type Food struct {
	ID            primitive.ObjectID
	Name          string
	Protein       uint16
	Carbohydrates uint16
	Fiber         uint16
	Sugar         uint16
	Fat           uint16
	Cholesterol   uint16
	Sodium        uint16
	Potassium     uint16
	VitaminA      uint16
	VitaminC      uint16
	Calcium       uint16
	Iron          uint16
}

type DietWindow struct {
	Date       time.Time
	Hour       uint8
	TimePeriod TimePeriod
	Entries    []DietEntry
}

type TimePeriod string

const (
	AM string = "AM"
	PM string = "PM"
)
