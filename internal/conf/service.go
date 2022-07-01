package conf

// Service file config
type Service struct {
	Name   string        `yaml:"name"`
	Build  ServiceBuild  `yaml:"build"`
	Deploy ServiceDeploy `yaml:"deploy"`
}

// ServiceBuild part of service file config
type ServiceBuild struct {
	Dockerfile string `yaml:"dockerfile"`
}

// ServiceDeploy part of service file config
type ServiceDeploy struct {
	// Template is the name of the template to use for the deployment
	Template string `yaml:"template"`

	// SecretMounts are used to mount secrets into the container
	SecretMounts []*SecretMount `yaml:"secretMounts"`

	// Config is used to setup environment variables into the container
	Config []*ServiceConfigItem `yaml:"config"`
}

// ServiceConfigItem is a single config item
type ServiceConfigItem struct {
	// Name of the config item
	Name string `yaml:"name"`

	// Hard coded values for this config
	Values map[string]string `yaml:"values"`
}

// SecretMount is a single secret mount
type SecretMount struct {
	MountName  string `yaml:"mountName"`
	SecretName string `yaml:"secretName"`
	MountPath  string `yaml:"mountPath"`
}
