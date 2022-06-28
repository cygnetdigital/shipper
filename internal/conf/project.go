package conf

// Project file config
type Project struct {
	Name           string        `yaml:"name"`
	Repo           string        `yaml:"repo"`
	RegistryPrefix string        `yaml:"registryPrefix"`
	Paths          []string      `yaml:"paths"`
	Gitops         ProjectGitops `yaml:"gitops"`
}

// ProjectGitops part of config file
type ProjectGitops struct {
	Repo         string `yaml:"repo"`
	ManifestPath string `yaml:"manifestPath"`
	TemplatePath string `yaml:"templatePath"`
	Namespace    string `yaml:"namespace"`
}
