package collector

func MakeNoContainerPoint() Point {
	return Point{
		MilliCpuUsage: 0,
		MemoryTotalMb: 0,
		MemoryRssMb:   0,
		MemoryLimitMb: 0,
		Running:       false,
	}
}
