package github

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/cygnetdigital/shipper/internal/destination"
)

// Get the destination state
func (s *Github) Get(ctx context.Context, projectName string) (*destination.Destination, error) {
	if projectName != s.projName {
		return nil, fmt.Errorf("project %s not supported", projectName)
	}

	_, rootPath, err := s.clone()
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer os.RemoveAll(rootPath)

	services, err := destination.LoadServices(path.Join(rootPath, s.bundlePath))
	if err != nil {
		return nil, fmt.Errorf("failed to load k8s manifests: %w", err)
	}

	return &destination.Destination{
		ProjectName: projectName,
		Services:    services.FilterByProject(s.projName),
	}, nil
}
