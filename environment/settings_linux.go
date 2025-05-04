package environment

import (
	"github.com/docker/docker/api/types/container"
)

func (l Limits) AsContainerResources() container.Resources {
	pids := l.ProcessLimit()

	return container.Resources{
		Memory:            l.BoundedMemoryLimit(),
		MemoryReservation: l.MemoryLimit * 1_000_000,
		MemorySwap:        l.ConvertedSwap(),
		CPUQuota:          l.ConvertedCpuLimit(),
		CPUPeriod:         100_000,
		CPUShares:         1024,
		BlkioWeight:       l.IoWeight,
		OomKillDisable:    &l.OOMDisabled,
		CpusetCpus:        l.Threads,
		PidsLimit:         &pids,
	}
}
