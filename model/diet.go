package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DietEntry struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	AddedAt time.Time          `json:"addedAt" bson:"addedAt" binding:"required"`
	Food    Food               `json:"food" bson:"food" binding:"required"`
}

type Food struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name" binding:"required"`
	Measurement   FoodMeasurement    `json:"measurement" bson:"measurement" binding:"required"`
	Calories      uint16             `json:"calories,omitempty" bson:"calories,omitempty"`
	Protein       uint16             `json:"protein,omitempty" bson:"calories,omitempty"`
	Carbohydrates uint16             `json:"carbohydrates,omitempty" bson:"carbohydrates,omitempty"`
	Fiber         uint16             `json:"fiber,omitempty" bson:"fiber,omitempty"`
	Sugar         uint16             `json:"sugar,omitempty" bson:"sugar,omitempty"`
	Fat           uint16             `json:"fat,omitempty" bson:"fat,omitempty"`
	Cholesterol   uint16             `json:"cholesterol,omitempty" bson:"cholesterol,omitempty"`
	Sodium        uint16             `json:"sodium,omitempty" bson:"calories,omitempty"`
	Potassium     uint16             `json:"potassium,omitempty" bson:"potassium,omitempty"`
	VitaminA      uint16             `json:"vitaminA,omitempty" bson:"vitaminA,omitempty"`
	VitaminC      uint16             `json:"vitaminc,omitempty" bson:"vitaminc,omitempty"`
	Calcium       uint16             `json:"calcium,omitempty" bson:"calcium,omitempty"`
	Iron          uint16             `json:"iron,omitempty" bson:"iron,omitempty"`
}

type DietWindow struct {
	Date       time.Time   `json:"date"`
	Hour       uint8       `json:"hour"`
	TimePeriod TimePeriod  `json:"timePeriod"`
	Entries    []DietEntry `json:"entries"`
}

type TimePeriod string

const (
	AM string = "AM"
	PM string = "PM"
)
