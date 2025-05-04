package server

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/pterodactyl/wings/config"
)

func getContainerConfig(ip *InstallationProcess) *container.Config {
	return &container.Config{
		Hostname:     "installer",
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		OpenStdin:    true,
		Tty:          true,
		Cmd:          []string{ip.Script.Entrypoint, "C:\\Pterodactyl-Install\\install.ps1"},
		Image:        ip.Script.ContainerImage,
		Env:          ip.Server.GetEnvironmentVariables(),
		Labels: map[string]string{
			"Service":       "Pterodactyl",
			"ContainerType": "server_installer",
		},
		User: "NT Authority\\System",
	}
}

func getContainerHostConfig(ip *InstallationProcess) *container.HostConfig {
	tmpfsSize := strconv.Itoa(int(config.Get().Docker.TmpfsSize))

	return &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Target:   "/Pterodactyl-Server",
				Source:   ip.Server.Filesystem().Path(),
				Type:     mount.TypeBind,
				ReadOnly: false,
			},
			{
				Target:   "/Pterodactyl-Install",
				Source:   ip.tempDir(),
				Type:     mount.TypeBind,
				ReadOnly: false,
			},
		},
		Resources: ip.resourceLimits(),
		Tmpfs: map[string]string{
			"/tmp": "rw,exec,nosuid,size=" + tmpfsSize + "M",
		},
		DNS: config.Get().Docker.Network.Dns,
		LogConfig: container.LogConfig{
			Type: "local",
			Config: map[string]string{
				"max-size": "5m",
				"max-file": "1",
				"compress": "false",
			},
		},
		Privileged:  false,
		NetworkMode: container.NetworkMode(config.Get().Docker.Network.Mode),
	}
}
