package receiver

type MeasureCommit struct {
	InstantSpeed float64
	AverageSpeed float64
}

func NewMeasureCommit(instantSpeed, averageSpeed float64) *MeasureCommit {
	return &MeasureCommit{
		InstantSpeed: instantSpeed,
		AverageSpeed: averageSpeed,
	}
}
