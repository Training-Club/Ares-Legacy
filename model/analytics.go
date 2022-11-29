package model

import (
	"fmt"
	"time"
)

type ExerciseDataCollection struct {
	Entries []ExerciseData `json:"entries" bson:"entries" binding:"required"`
}

type ExerciseDataWeight struct {
	Total            float32
	HighestWeightSet Exercise
}

type ExerciseDataReps struct {
	Total            float32
	HighestVolumeSet Exercise
}

type ExerciseDataTime struct {
	Total      float32
	FastestSet Exercise
	LongestSet Exercise
}

type ExerciseDataDistance struct {
	Total      float32
	LongestSet Exercise
}

type ExerciseData struct {
	Exercises []Exercise           `json:"exercises" bson:"exercises" binding:"required"`
	Weight    ExerciseDataWeight   `json:"weight,omitempty" bson:"weight,omitempty"`
	Reps      ExerciseDataReps     `json:"reps,omitempty" bson:"reps,omitempty"`
	Time      ExerciseDataTime     `json:"time,omitempty" bson:"time,omitempty"`
	Distance  ExerciseDataDistance `json:"distance,omitempty" bson:"distance,omitempty"`
}

type QueryPeriod string

const (
	WEEK        QueryPeriod = "WEEK"
	MONTH       QueryPeriod = "MONTH"
	THREE_MONTH QueryPeriod = "THREE_MONTH"
	SIX_MONTH   QueryPeriod = "SIX_MONTH"
	YEAR        QueryPeriod = "YEAR"
	ALL         QueryPeriod = "ALL"
)

// GetQueryPeriod accepts a string value and converts it
// in to a QueryPeriod enum
func GetQueryPeriod(value string) (QueryPeriod, error) {
	switch value {
	case "WEEK":
		return WEEK, nil
	case "MONTH":
		return MONTH, nil
	case "THREE_MONTH":
		return THREE_MONTH, nil
	case "SIX_MONTH":
		return SIX_MONTH, nil
	case "YEAR":
		return YEAR, nil
	case "ALL":
		return ALL, nil
	}

	return ALL, fmt.Errorf("invalid query period: " + value)
}

// GetQueryPeriodDuration accepts a QueryPeriod and returns the current date
// deducted from the specified time format.
func GetQueryPeriodDuration(period QueryPeriod) time.Time {
	defaultTime := time.Now()

	switch period {
	case WEEK:
		defaultTime.AddDate(0, 0, -7)
		break
	case MONTH:
		defaultTime.AddDate(0, -1, 0)
		break
	case THREE_MONTH:
		defaultTime.AddDate(0, -3, 0)
		break
	case SIX_MONTH:
		defaultTime.AddDate(0, -6, 0)
		break
	case YEAR:
		defaultTime.AddDate(-1, 0, 0)
		break
	default:
		defaultTime.AddDate(-30, 0, 0)
	}

	return defaultTime
}
