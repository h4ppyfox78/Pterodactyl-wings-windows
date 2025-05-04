package docker

import (
	"math"

	"github.com/docker/docker/api/types"
)

// The "docker stats" CLI call does not return the same value as the types.MemoryStats.Usage
// value which can be rather confusing to people trying to compare panel usage to
// their stats output.
//
// This math is from their CLI repository in order to show the same values to avoid people
// bothering me about it. It should also reflect a slightly more correct memory value anyways.
//
// @see https://github.com/docker/cli/blob/96e1d1d6/cli/command/container/stats_helpers.go#L227-L249
func calculateDockerMemory(stats types.MemoryStats) uint64 {
	if v, ok := stats.Stats["total_inactive_file"]; ok && v < stats.Usage {
		return stats.Usage - v
	}

	if v := stats.Stats["inactive_file"]; v < stats.Usage {
		return stats.Usage - v
	}

	return stats.Usage
}

// Calculates the absolute CPU usage used by the server process on the system, not constrained
// by the defined CPU limits on the container.
//
// @see https://github.com/docker/cli/blob/aa097cf1aa19099da70930460250797c8920b709/cli/command/container/stats_helpers.go#L166
func calculateDockerAbsoluteCpu(v types.StatsJSON) float64 {
	pStats := v.PreCPUStats
	stats := v.CPUStats
	// Calculate the change in CPU usage between the current and previous reading.
	cpuDelta := float64(stats.CPUUsage.TotalUsage) - float64(pStats.CPUUsage.TotalUsage)

	// Calculate the change for the entire system's CPU usage between current and previous reading.
	systemDelta := float64(stats.SystemUsage) - float64(pStats.SystemUsage)

	// Calculate the total number of CPU cores being used.
	cpus := float64(stats.OnlineCPUs)
	if cpus == 0.0 {
		cpus = float64(len(stats.CPUUsage.PercpuUsage))
	}

	percent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		percent = (cpuDelta / systemDelta) * cpus * 100.0
	}

	return math.Round(percent*1000) / 1000
}
