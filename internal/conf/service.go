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
	Template string `yaml:"template"`
}
