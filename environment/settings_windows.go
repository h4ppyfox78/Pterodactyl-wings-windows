package environment

import (
	"github.com/docker/docker/api/types/container"
)

func (l Limits) AsContainerResources() container.Resources {
	return container.Resources{
		Memory:     l.BoundedMemoryLimit(),
		CPUQuota:   l.ConvertedCpuLimit(),
		CPUShares:  1024,
		CpusetCpus: l.Threads,
	}
}
