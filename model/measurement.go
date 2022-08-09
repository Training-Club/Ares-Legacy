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

type FoodMeasurement string

const (
	OUNCE    string = "OUNCE"
	GRAM     string = "GRAM"
	POUND    string = "POUND"
	KILOGRAM string = "KILOGRAM"
	CUP      string = "CUP"
	SERVING  string = "SERVING"
)
