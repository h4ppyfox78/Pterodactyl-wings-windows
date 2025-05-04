package server

import (
	"github.com/pterodactyl/wings/environment"
)

// Returns the default container mounts for the server instance. This includes the data directory
// for the server. Previously this would also mount in host timezone files, however we've moved from
// that approach to just setting `TZ=Timezone` environment values in containers which should work
// in most scenarios.
func (s *Server) Mounts() []environment.Mount {
	m := []environment.Mount{
		{
			Default:  true,
			Target:   "/Container",
			Source:   s.Filesystem().Path(),
			ReadOnly: false,
		},
	}

	// Also include any of this server's custom mounts when returning them.
	return append(m, s.customMounts()...)
}
