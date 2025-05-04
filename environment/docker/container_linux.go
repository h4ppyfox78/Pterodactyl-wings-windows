package docker

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/daemon/logger/local"
	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/environment"
)

// getContainerUser gets the user for the container
func getContainerUser() string {
	return strconv.Itoa(config.Get().System.User.Uid) + ":" + strconv.Itoa(config.Get().System.User.Gid)
}

func getContainerHostConfig(e *Environment, a environment.Allocations) *container.HostConfig {
	tmpfsSize := strconv.Itoa(int(config.Get().Docker.TmpfsSize))

	return &container.HostConfig{
		PortBindings: a.DockerBindings(),

		// Configure the mounts for this container. First mount the server data directory
		// into the container as an r/w bind.
		Mounts: e.convertMounts(),

		// Configure the /tmp folder mapping in containers. This is necessary for some
		// games that need to make use of it for downloads and other installation processes.
		Tmpfs: map[string]string{
			"/tmp": "rw,exec,nosuid,size=" + tmpfsSize + "M",
		},

		// Define resource limits for the container based on the data passed through
		// from the Panel.
		Resources: e.Configuration.Limits().AsContainerResources(),

		DNS: config.Get().Docker.Network.Dns,

		// Configure logging for the container to make it easier on the Daemon to grab
		// the server output. Ensure that we don't use too much space on the host machine
		// since we only need it for the last few hundred lines of output and don't care
		// about anything else in it.
		LogConfig: container.LogConfig{
			Type: local.Name,
			Config: map[string]string{
				"max-size": "5m",
				"max-file": "1",
				"compress": "false",
				"mode":     "non-blocking",
			},
		},

		SecurityOpt:    []string{"no-new-privileges"},
		ReadonlyRootfs: true,
		CapDrop: []string{
			"setpcap", "mknod", "audit_write", "net_raw", "dac_override",
			"fowner", "fsetid", "net_bind_service", "sys_chroot", "setfcap",
		},
		NetworkMode: container.NetworkMode(config.Get().Docker.Network.Mode),
	}
}

func (e *Environment) resources() container.Resources {
	l := e.Configuration.Limits()
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
