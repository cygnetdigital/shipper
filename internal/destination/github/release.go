package github

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/cygnetdigital/shipper/internal/destination"
)

// Release to the destination
func (s *Github) Release(ctx context.Context, p *destination.ReleaseParams) (*destination.ReleaseResp, error) {
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
		return nil, fmt.Errorf("version %s is already active", p.Version)
	}

	slug, err := destination.SlugifyServiceName(svc.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to slugify service name: %w", err)
	}

	args := &destination.ReleaseContext{
		Project:         svc.Project,
		Name:            svc.Name,
		Version:         p.Version,
		SlugName:        slug,
		SlugNameVersion: fmt.Sprintf("%s-%s", slug, p.Version),
		Namespace:       s.namespace,
	}

	manifestRoot := path.Join(rootPath, s.bundlePath)

	if err := destination.WriteReleaseBundle(manifestRoot, args); err != nil {
		return nil, fmt.Errorf("failed to write bundle: %w", err)
	}

	msg := fmt.Sprintf("Releasing %s/%s/%s", p.Project, p.Service, p.Version)

	hash, err := s.commit(msg, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &destination.ReleaseResp{Hash: hash}, nil
}
