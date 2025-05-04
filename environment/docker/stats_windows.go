package docker

import (
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
	return stats.PrivateWorkingSet
}

// Calculates the absolute CPU usage used by the server process on the system, not constrained
// by the defined CPU limits on the container.
//
// @see https://github.com/docker/cli/blob/aa097cf1aa19099da70930460250797c8920b709/cli/command/container/stats_helpers.go#L166
func calculateDockerAbsoluteCpu(v types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}
