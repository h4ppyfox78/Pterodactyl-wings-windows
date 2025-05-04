package config

type DockerNetworkConfiguration struct {
	// The interface that should be used to create the network. Must not conflict
	// with any other interfaces in use by Docker or on the system.
	Interface string `default:"172.18.0.1" json:"interface" yaml:"interface"`

	// The DNS settings for containers.
	Dns []string `default:"[\"1.1.1.1\", \"1.0.0.1\"]"`

	// The name of the network to use. If this network already exists it will not
	// be created. If it is not found, a new network will be created using the interface
	// defined.
	Name       string                  `default:"pterodactyl_nw"`
	ISPN       bool                    `default:"false" yaml:"ispn"`
	Driver     string                  `default:"nat"`
	Mode       string                  `default:"pterodactyl_nw" yaml:"network_mode"`
	IsInternal bool                    `default:"false" yaml:"is_internal"`
	EnableICC  bool                    `default:"true" yaml:"enable_icc"`
	Interfaces dockerNetworkInterfaces `yaml:"interfaces"`
}
