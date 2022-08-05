package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	SessionName string             `json:"sessionName,omitempty" bson:"sessionName,omitempty" binding:"required"`
	Author      primitive.ObjectID `json:"author,omitempty" bson:"author,omitempty" binding:"required"`
	Status      SessionStatus      `json:"status,omitempty" bson:"status,omitempty" binding:"required"`
	Timestamp   time.Time          `json:"timestamp,omitempty" bson:"timestamp,omitempty" time-format:"" binding:"required"`
	Exercises   []Exercise         `json:"exercises,omitempty" bson:"exercises,omitempty" binding:"required"`
}

type Exercise struct {
	ExerciseName       string               `json:"exerciseName,omitempty" bson:"exerciseName,omitempty" binding:"required"`
	AddedAt            time.Time            `json:"addedAt,omitempty" bson:"addedAt,omitempty" binding:"required"`
	Values             ExerciseValue        `json:"values,omitempty" bson:"values,omitempty" binding:"required"`
	Type               ExerciseType         `json:"type,omitempty" bson:"type,omitempty" binding:"required"`
	AdditionalExercise []AdditionalExercise `json:"additionalExercises,omitempty" bson:"additionalExercises,omitempty"`
}

type AdditionalExercise struct {
	ExerciseName string                 `json:"exerciseName" bson:"exerciseName,omitempty" binding:"required"`
	AddedAt      time.Time              `json:"addedAt,omitempty" bson:"addedAt,omitempty" binding:"required"`
	Values       ExerciseValue          `json:"values,omitempty" bson:"values,omitempty" binding:"required"`
	Type         AdditionalExerciseType `json:"type,omitempty" bson:"type,omitempty" binding:"required"`
}

type ExerciseValue struct {
	Reps     uint8                 `json:"reps,omitempty" bson:"reps,omitempty"`
	Weight   ExerciseValueWeight   `json:"weight,omitempty" bson:"weight,omitempty"`
	Distance ExerciseValueDistance `json:"distance,omitempty" bson:"distance,omitempty"`
	Time     ExerciseValueTime     `json:"time,omitempty" bson:"time,omitempty"`
}

type ExerciseValueWeight struct {
	Value               float32           `json:"weightValue,omitempty" bson:"weightValue,omitempty" binding:"required"`
	Measurement         MeasurementSystem `json:"weightMeasurementSystem,omitempty" bson:"weightMeasurementSystem,omitempty" binding:"required"`
	PlateCounterEnabled bool              `json:"plateCounterEnabled,omitempty" bson:"plateCounterEnabled,omitempty"`
}

type ExerciseValueDistance struct {
	Value       uint32              `json:"distanceValue,omitempty" bson:"distanceValue,omitempty" binding:"required"`
	Measurement DistanceMeasurement `json:"distanceMeasurementSystem,omitempty" bson:"distanceMeasurementSystem,omitempty" binding:"required"`
}

type ExerciseValueTime struct {
	Value            uint32 `json:"timeValue,omitempty" bson:"timeValue,omitempty" binding:"required"`
	ShowMilliseconds bool   `json:"timeRenderMillis,omitempty" bson:"timeRenderMillis,omitempty" binding:"required"`
}

type ExerciseType string
type AdditionalExerciseType string
type SessionStatus string

const (
	WEIGHTED_REPS ExerciseType = "weighted_reps"
	WEIGHTED_TIME ExerciseType = "weighted_time"
	DISTANCE_TIME ExerciseType = "distance_time"
	REPS          ExerciseType = "reps"
	TIME          ExerciseType = "time"
	DISTANCE                   = "distance"
)

const (
	SUPERSET AdditionalExerciseType = "superset"
	DROPSET  AdditionalExerciseType = "dropset"
)

const (
	DRAFT       SessionStatus = "draft"
	IN_PROGRESS SessionStatus = "in_progress"
	ASSIGNED                  = "assigned"
	COMPLETED                 = "completed"
)
