package collector

import (
	"time"
)

type Point struct {
	CpuUsage      float64 // TODO
	MemoryTotalMb uint64
	MemoryRssMb   uint64
	MemoryLimitMb uint64
	Running       bool
}

type State struct {
	Time                time.Time
	AccumulatedCpuUsage uint64
}

type Collector interface {
	GetPoint(lastState State) (Point, State)
}
