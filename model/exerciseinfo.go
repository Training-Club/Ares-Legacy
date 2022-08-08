package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ExerciseInfo struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name,omitempty" bson:"name,omitempty" binding:"required"`
	Verified     bool               `json:"verified,omitempty" bson:"verified,omitempty"`
	Type         ExerciseType       `json:"type,omitempty" bson:"type,omitempty" binding:"required"`
	MuscleGroups []MuscleGroup      `json:"muscleGroups,omitempty" bson:"muscleGroups,omitempty"`
	Equipment    ExerciseEquipment  `json:"equipment,omitempty"`
}

type ExerciseEquipment string
type MuscleGroup string

const (
	BARBELL    ExerciseEquipment = "BARBELL"
	DUMBBELL   ExerciseEquipment = "DUMBBELL"
	MACHINE    ExerciseEquipment = "MACHINE"
	KETTLEBELL ExerciseEquipment = "KETTLEBELL"
)

const (
	NECK       MuscleGroup = "NECK"
	SHOULDERS  MuscleGroup = "SHOULDERS"
	UPPER_ARMS MuscleGroup = "UPPER_ARMS"
	FOREARMS   MuscleGroup = "FOREARMS"
	BACK       MuscleGroup = "BACK"
	CHEST      MuscleGroup = "CHEST"
	THIGHS     MuscleGroup = "THIGHS"
	CALVES     MuscleGroup = "CALVES"
)
