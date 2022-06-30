package handler

import (
	"context"
	"fmt"

	"github.com/cygnetdigital/shipper/internal/destination"
)

// ReleaseParams describe the Release we wish to perform
type ReleaseParams struct {
	// Name of the project
	Project string

	// Service to release
	Service string

	// Version of the service to release
	Version string

	// Confirm should be true to actually do the deploy
	Confirm bool
}

// ReleaseResp is the result from a Release
type ReleaseResp struct {
	Project string
	Service string
	Version string
	Done    bool
}

// Release ...
func (h *LocalHandler) Release(ctx context.Context, p *ReleaseParams) (*ReleaseResp, error) {
	if p.Confirm {
		relreq := &destination.ReleaseParams{
			Project: p.Project,
			Service: p.Service,
			Version: p.Version,
		}

		if _, err := h.Dest.Release(ctx, relreq); err != nil {
			return nil, fmt.Errorf("failed to release destination: %w", err)
		}

		return &ReleaseResp{Done: true}, nil
	}

	dest, err := h.Dest.Get(ctx, p.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	svc := dest.Services.Lookup(p.Service)
	if svc == nil {
		return nil, fmt.Errorf("service not found")
	}

	if p.Version == "" {
		if len(svc.Deploys) == 0 {
			return nil, fmt.Errorf("no deploys found for service")
		}

		p.Version = svc.Deploys[len(svc.Deploys)-1].Version
	}

	if svc.CurrentReleaseVersion == p.Version {
		return &ReleaseResp{
			Project: p.Project,
			Service: p.Service,
			Version: p.Version,
			Done:    true,
		}, nil
	}

	if !svc.HasVersion(p.Version) {
		return nil, fmt.Errorf("version %s not found", p.Version)
	}

	return &ReleaseResp{
		Project: svc.Project,
		Service: svc.Name,
		Version: p.Version,
	}, nil
}
