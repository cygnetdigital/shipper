package handler

import (
	"context"
	"fmt"

	"github.com/cygnetdigital/shipper/internal/destination"
)

// RemoveParams describe the Remove we wish to perform
type RemoveParams struct {
	// Name of the project
	Project string

	// Service to remove
	Service string

	// Version of the service to remove
	Version string

	// Confirm should be true to actually do the deploy
	Confirm bool
}

// RemoveResp is the result from a Remove
type RemoveResp struct {
	Project string
	Service string
	Version string
	Done    bool
}

// Remove ...
func (h *LocalHandler) Remove(ctx context.Context, p *RemoveParams) (*RemoveResp, error) {

	if p.Service == "" || p.Version == "" {
		return nil, fmt.Errorf("service and version are required")
	}

	if p.Confirm {
		remreq := &destination.RemoveParams{
			Project: p.Project,
			Service: p.Service,
			Version: p.Version,
		}

		if _, err := h.Dest.Remove(ctx, remreq); err != nil {
			return nil, fmt.Errorf("failed to remove from destination: %w", err)
		}

		return &RemoveResp{Done: true}, nil
	}

	dest, err := h.Dest.Get(ctx, p.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination: %w", err)
	}

	svc := dest.Services.Lookup(p.Service)
	if svc == nil {
		return nil, fmt.Errorf("service not found")
	}

	if svc.CurrentReleaseVersion == p.Version {
		return nil, fmt.Errorf("cannot remove current release version")
	}

	if !svc.HasVersion(p.Version) {
		return nil, fmt.Errorf("version %s not found", p.Version)
	}

	return &RemoveResp{
		Project: svc.Project,
		Service: svc.Name,
		Version: p.Version,
	}, nil
}
