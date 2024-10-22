package receiver

type Measure struct {
	AverageSpeed float64
	InstantSpeed float64
}

func NewMeasure(averageSpeed, instantSpeed float64) Measure {
	return Measure{
		AverageSpeed: averageSpeed,
		InstantSpeed: instantSpeed,
	}
}
