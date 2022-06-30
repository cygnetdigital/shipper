package github

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/cygnetdigital/shipper/internal/destination"
)

// Remove from the destination
func (s *Github) Remove(ctx context.Context, p *destination.RemoveParams) (*destination.RemoveResp, error) {
	if p.Project != s.projName {
		return nil, fmt.Errorf("project %s not supported", p.Project)
	}

	repo, rootPath, err := s.clone()
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer os.RemoveAll(rootPath)

	svcs, err := destination.LoadServices(path.Join(rootPath, s.bundlePath))
	if err != nil {
		return nil, fmt.Errorf("failed to load k8s manifests: %w", err)
	}

	svc := svcs.LookupByProjectAndName(p.Project, p.Service)
	if svc == nil {
		return nil, fmt.Errorf("service not found")
	}

	if !svc.HasVersion(p.Version) {
		return nil, fmt.Errorf("version %s not found", p.Version)
	}

	if svc.CurrentReleaseVersion == p.Version {
		return nil, fmt.Errorf("version %s is active so it cannot be removed", p.Version)
	}

	manifestRoot := path.Join(rootPath, s.bundlePath)

	if err := destination.DeleteDeployBundle(manifestRoot, svc.Name, p.Version); err != nil {
		return nil, fmt.Errorf("failed to delete deploy: %w", err)
	}

	msg := fmt.Sprintf("Removing %s/%s/%s", p.Project, p.Service, p.Version)

	hash, err := s.commit(msg, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &destination.RemoveResp{Hash: hash}, nil
}
