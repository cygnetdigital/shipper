package destination

// Services ...
type Services []*Service

// FilterByProject name
func (s Services) FilterByProject(proj string) (out Services) {
	for _, service := range s {
		if service.ProjectName == proj {
			out = append(out, service)
		}
	}

	return out
}

// Lookup a service by name
func (s Services) Lookup(name string) *Service {
	for _, service := range s {
		if service.ServiceName == name {
			return service
		}
	}

	return nil
}

// NextDeployVersionFor returns the next deploy version for the given service
func (s Services) NextDeployVersionFor(name string) int {
	service := s.Lookup(name)
	if service == nil {
		return 1
	}

	return service.NextDeployVersion()
}

// CurrentDeployVersionFor returns the deploy version for the given service
func (s Services) CurrentDeployVersionFor(name string) int {
	service := s.Lookup(name)
	if service == nil {
		return 0
	}

	return service.DeployVersion()
}
