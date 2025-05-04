package environment

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pterodactyl/wings/config"
)

// Creates a new network on the machine if one does not exist already.
func createDockerNetwork(ctx context.Context, cli *client.Client) error {
	nw := config.Get().Docker.Network
	_, err := cli.NetworkCreate(ctx, nw.Name, types.NetworkCreate{
		Driver:     nw.Driver,
		EnableIPv6: true,
		Internal:   nw.IsInternal,
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{{
				Subnet:  nw.Interfaces.V4.Subnet,
				Gateway: nw.Interfaces.V4.Gateway,
			}, {
				Subnet:  nw.Interfaces.V6.Subnet,
				Gateway: nw.Interfaces.V6.Gateway,
			}},
		},
		Options: map[string]string{
			"encryption": "false",
			"com.docker.network.bridge.default_bridge":       "false",
			"com.docker.network.bridge.enable_icc":           strconv.FormatBool(nw.EnableICC),
			"com.docker.network.bridge.enable_ip_masquerade": "true",
			"com.docker.network.bridge.host_binding_ipv4":    "0.0.0.0",
			"com.docker.network.bridge.name":                 "pterodactyl0",
			"com.docker.network.driver.mtu":                  "1500",
		},
	})
	if err != nil {
		return err
	}
	if nw.Driver != "host" && nw.Driver != "overlay" && nw.Driver != "weavemesh" {
		config.Update(func(c *config.Configuration) {
			c.Docker.Network.Interface = c.Docker.Network.Interfaces.V4.Gateway
		})
	}
	return nil
}
