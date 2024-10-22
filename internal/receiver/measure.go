package receiver

type Measure interface {
	AddReadedLen(length int64)
	Commit() *MeasureCommit
}
