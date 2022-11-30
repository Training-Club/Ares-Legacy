package controller

import (
	"ares/database"
	"ares/model"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
)

// GetRecordsByExercise is an experimental analytics feature to query
// user stats in a graph-able format
func (controller *AresController) GetRecordsByExercise() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// requestAccountId := ctx.GetString("accountId")

		accountId := ctx.Param("accountId")
		exerciseName := ctx.Param("exerciseName")
		queryPeriod := ctx.DefaultQuery("period", "YEAR")
		page := ctx.DefaultQuery("page", "0")

		accountIdHex, err := primitive.ObjectIDFromHex(accountId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad account id hex"})
			return
		}

		period, err := model.GetQueryPeriod(queryPeriod)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad query period"})
			return
		}
		before := model.GetQueryPeriodDuration(period)

		pageNumber, err := strconv.Atoi(page)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid page number"})
			return
		}

		filter := bson.M{}
		filter["author"] = accountIdHex
		filter["exercises.exerciseName"] = bson.M{"$in": exerciseName}
		filter["timestamp"] = bson.M{"$gte": primitive.NewDateTimeFromTime(before)}

		sessions, err := database.FindManyDocumentsByFilterWithOpts[model.Session](database.QueryParams{
			MongoClient:    controller.DB,
			DatabaseName:   controller.DatabaseName,
			CollectionName: controller.CollectionName,
		}, filter, options.Find().SetLimit(500).SetSkip(int64(pageNumber*500)).SetSort(bson.D{{Key: "timestamp", Value: -1}}))

		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "no sessions found"})
				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to query sessions: " + err.Error()})
			return
		}

		var exerciseData []model.ExerciseData
		for _, session := range sessions {
			var exercises []model.Exercise
			var totalVolume float32
			var totalSets float32
			var totalReps float32
			var totalTime float32
			var totalDistance float32
			var highestVolumeSet model.Exercise
			var highestWeightSet model.Exercise
			var longestDistanceSet model.Exercise
			var shortestTimeSet model.Exercise
			var longestTimeSet model.Exercise

			for _, exercise := range session.Exercises {
				if exercise.ExerciseName != exerciseName {
					continue
				}

				// add to overall array
				exercises = append(exercises, exercise)

				// default reps to 1
				reps := exercise.Values.Reps
				if reps <= 0 {
					reps = 1
				}

				totalSets += 1
				totalReps += float32(reps)

				// assign highest volume, weight, and total volume
				if exercise.Values.Weight.Value > 0 {
					volume := exercise.Values.Weight.Value * float32(reps)

					totalVolume += volume

					// highest volume
					currentHighestVolume := highestVolumeSet.Values.Weight.Value * float32(highestVolumeSet.Values.Reps)
					if volume > currentHighestVolume {
						highestVolumeSet = exercise
					}

					// highest weight
					if exercise.Values.Weight.Value > highestWeightSet.Values.Weight.Value {
						highestWeightSet = exercise
					}
				}

				// assign fastest and slowest sets and total time
				if exercise.Values.Time.Value > 0 {
					totalTime += float32(exercise.Values.Time.Value)

					if shortestTimeSet.Values.Time.Value > exercise.Values.Time.Value {
						shortestTimeSet = exercise
					}

					if longestTimeSet.Values.Time.Value < exercise.Values.Time.Value {
						longestTimeSet = exercise
					}
				}

				// assign longest distance and total distance
				if exercise.Values.Distance.Value > 0 {
					totalDistance += float32(exercise.Values.Distance.Value)

					if longestDistanceSet.Values.Distance.Value < exercise.Values.Distance.Value {
						longestDistanceSet = exercise
					}
				}
			}

			data := model.ExerciseData{
				Exercises: exercises,
				Weight: model.ExerciseDataWeight{
					Total:            totalVolume,
					HighestWeightSet: highestWeightSet,
				},
				Reps: model.ExerciseDataReps{
					Total:            totalReps,
					HighestVolumeSet: highestVolumeSet,
				},
				Time: model.ExerciseDataTime{
					Total:      totalTime,
					FastestSet: shortestTimeSet,
					LongestSet: longestTimeSet,
				},
				Distance: model.ExerciseDataDistance{
					Total:      totalDistance,
					LongestSet: longestDistanceSet,
				},
			}

			exerciseData = append(exerciseData, data)
		}
	}
}
