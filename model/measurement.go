package model

type MeasurementSystem string

const (
	IMPERIAL MeasurementSystem = "imperial"
	METRIC   MeasurementSystem = "metric"
)

type DistanceMeasurement string

const (
	METER     DistanceMeasurement = "meter"
	MILE      DistanceMeasurement = "mile"
	FEET      DistanceMeasurement = "feet"
	KILOMETER DistanceMeasurement = "kilometer"
)
