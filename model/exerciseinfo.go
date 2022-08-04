package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ExerciseInfo struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name,omitempty" bson:"name,omitempty" binding:"required"`
	Verified     bool               `json:"verified,omitempty" bson:"verified,omitempty"`
	MuscleGroups []MuscleGroup      `json:"muscleGroups,omitempty" bson:"muscleGroups,omitempty"`
	Equipment    ExerciseEquipment  `json:"equipment,omitempty"`
}

type ExerciseEquipment string
type MuscleGroup string

const (
	BARBELL    ExerciseEquipment = "barbell"
	DUMBBELL   ExerciseEquipment = "dumbbell"
	MACHINE    ExerciseEquipment = "machine"
	KETTLEBELL ExerciseEquipment = "kettlebell"
)

const (
	CALVES     MuscleGroup = "calves"
	HAMSTRING  MuscleGroup = "hamstring"
	QUADS      MuscleGroup = "quads"
	GLUTES     MuscleGroup = "glutes"
	BICEPS     MuscleGroup = "biceps"
	TRICEPS    MuscleGroup = "triceps"
	TRAPS      MuscleGroup = "traps"
	UPPER_BACK MuscleGroup = "upperBack"
	LOWER_BACK MuscleGroup = "lowerBack"
	CHEST      MuscleGroup = "chest"
	ABS        MuscleGroup = "abs"
)
