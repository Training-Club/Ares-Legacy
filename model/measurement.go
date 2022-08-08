package model

type MeasurementSystem string

const (
	IMPERIAL MeasurementSystem = "IMPERIAL"
	METRIC   MeasurementSystem = "METRIC"
)

type DistanceMeasurement string

const (
	METER     DistanceMeasurement = "METER"
	MILE      DistanceMeasurement = "MILE"
	FEET      DistanceMeasurement = "FEET"
	KILOMETER DistanceMeasurement = "KILOMETER"
)
