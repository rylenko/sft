package receiver

import (
	"sync"
	"time"
)

type SyncMeasure struct {
	startTime time.Time
	lastTime time.Time

	totalReadedLen int64
	instantReadedLen int64

	mutex sync.Mutex
}

func (measure *SyncMeasure) AddReadedLen(length int64) {
	measure.mutex.Lock()
	defer measure.mutex.Unlock()

	measure.instantReadedLen += length
	measure.totalReadedLen += length
}

func (measure *SyncMeasure) Commit() *MeasureCommit {
	measure.mutex.Lock()
	defer measure.mutex.Unlock()

	currentTime := time.Now()

	// Get instant speed using instant readed length and instant elapsed seconds.
	instantElapsedSeconds := currentTime.Sub(measure.lastTime).Seconds()
	if instantElapsedSeconds <= 0 {
		instantElapsedSeconds = 0.0001
	}
	instantSpeed := float64(measure.instantReadedLen) / instantElapsedSeconds

	// Get average speed using total readed length and total elapsed seconds.
	totalElapsedSeconds := currentTime.Sub(measure.startTime).Seconds()
	if totalElapsedSeconds <= 0 {
		totalElapsedSeconds = 0.0001
	}
	averageSpeed := float64(measure.totalReadedLen) / totalElapsedSeconds

	// Update last measure time and instant readed length.
	measure.lastTime = currentTime
	measure.instantReadedLen = 0

	return NewMeasureCommit(instantSpeed, averageSpeed)
}

func NewSyncMeasure() *SyncMeasure {
	currentTime := time.Now()

	return &SyncMeasure{
		startTime: currentTime,
		lastTime: currentTime,
		totalReadedLen: 0,
		instantReadedLen: 0,
	}
}
