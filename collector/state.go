package collector

import (
	"time"
)

func MakeNoContainerState() State {
	return State{
		Time:                time.Now(),
		AccumulatedCpuUsage: MaxUint64,
	}
}
