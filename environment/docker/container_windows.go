package docker

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/daemon/logger/local"
	"github.com/docker/go-connections/nat"
	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/environment"
)

// getContainerUser gets the user for the container
func getContainerUser() string {
	return config.Get().System.Username
}

// getDockerBindingsForWindows As Windows does not support the IP being set on NAT bindings, we will remap the mappings
// so the IP is not included
func getDockerBindingsForWindows(a environment.Allocations) nat.PortMap {
	portBindings := a.DockerBindings()

	for p, binds := range portBindings {
		for i, alloc := range binds {
			portBindings[p][i] = nat.PortBinding{
				HostIP:   "",
				HostPort: alloc.HostPort,
			}
		}
	}

	return portBindings
}

func getContainerHostConfig(e *Environment, a environment.Allocations) *container.HostConfig {
	tmpfsSize := strconv.Itoa(int(config.Get().Docker.TmpfsSize))

	return &container.HostConfig{
		PortBindings: getDockerBindingsForWindows(a),

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

		// This security opt `no-new-privileges` is not supported on Windows
		// I cannot find something simalar for Windows, also Windows doesn't have Sudo,
		// so maybe this can be safely ignored
		// SecurityOpt:    []string{"no-new-privileges"},
		ReadonlyRootfs: false,
		CapDrop: []string{
			"setpcap", "mknod", "audit_write", "net_raw", "dac_override",
			"fowner", "fsetid", "net_bind_service", "sys_chroot", "setfcap",
		},
		NetworkMode: container.NetworkMode(config.Get().Docker.Network.Mode),
	}
}

func (e *Environment) resources() container.Resources {
	l := e.Configuration.Limits()

	return container.Resources{
		Memory:     l.BoundedMemoryLimit(),
		CPUQuota:   l.ConvertedCpuLimit(),
		CPUShares:  1024,
		CpusetCpus: l.Threads,
	}
}
