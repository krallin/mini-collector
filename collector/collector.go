package collector

import (
	"fmt"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
	"time"
)

type subsystem interface {
	Name() string
	GetStats(path string, stats *cgroups.Stats) error
}

type collector struct {
	dockerName  string
	statsBuffer cgroups.Stats
	subsystems  []subsystem
}

func NewCollector(dockerName string) Collector {
	statsBuffer := *cgroups.NewStats()

	subsystems := []subsystem{
		&fs.CpuGroup{},
		&fs.MemoryGroup{},
		&fs.CpuacctGroup{},
	}

	return &collector{
		dockerName:  dockerName,
		statsBuffer: statsBuffer,
		subsystems:  subsystems,
	}
}


func (c *collector) GetPoint(lastState State) (Point, State) {
	for _, subsys := range c.subsystems {
		cgPath := fmt.Sprintf("/sys/fs/cgroup/%s/docker/%s", subsys.Name(), c.dockerName)

		err := subsys.GetStats(cgPath, &c.statsBuffer)
		if err != nil {
			// TODO: Logging
			fmt.Printf("%s.GetStats Error: %+v\n", subsys.Name(), err)
			return MakeNoContainerPoint(), MakeNoContainerState()
		}
	}

	pollTime := time.Now()

	accumulatedCpuUsage := c.statsBuffer.CpuStats.CpuUsage.TotalUsage

	var cpuUsage float64
	if accumulatedCpuUsage > lastState.AccumulatedCpuUsage {
		elapsedCpu := float64(accumulatedCpuUsage - lastState.AccumulatedCpuUsage)
		elapsedTime := float64(pollTime.Sub(lastState.Time).Nanoseconds())
		cpuUsage = elapsedCpu / elapsedTime
	} else {
		cpuUsage = 0
	}

	baseRssMemory := c.statsBuffer.MemoryStats.Stats["rss"]
	mappedFileMemory := c.statsBuffer.MemoryStats.Stats["mapped_file"]
	virtualMemory := c.statsBuffer.MemoryStats.Usage.Usage
	limitMemory := c.statsBuffer.MemoryStats.Usage.Limit

	point := Point{
		CpuUsage:      cpuUsage,
		MemoryTotalMb: virtualMemory / MbInBytes,
		MemoryRssMb:   (baseRssMemory + mappedFileMemory) / MbInBytes,
		MemoryLimitMb: (limitMemory) / MbInBytes,
		Running:       true,
	}

	state := State{
		Time:                pollTime,
		AccumulatedCpuUsage: accumulatedCpuUsage,
	}

	return point, state
}
