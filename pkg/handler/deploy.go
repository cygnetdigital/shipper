package handler

import (
	"context"
	"fmt"
	"path"

	"github.com/cygnetdigital/shipper/internal/destination"
	"github.com/cygnetdigital/shipper/internal/source"
)

// DeployParams describe the deploy we wish to perform
type DeployParams struct {
	// Name of the project
	ProjectName string

	// Ref can be a PR number or git branch
	Ref string

	// Confirm params should be provided if we are confirming a deploy
	Confirm *ConfirmDeployParams
}

// ConfirmDeployParams are the params required to perform the deploy
type ConfirmDeployParams struct {
	CommitHash source.GitHash
	Requests   []*ServiceDeployRequest
}

// DeployResp is the result from a deploy
type DeployResp struct {
	Source   *source.Source
	Services []*ServiceDeployStatus
	Complete bool
}

// ServiceDeployRequest ...
type ServiceDeployRequest struct {
	ServiceName   string
	DeployVersion int
}

// ServiceDeployStatus ...
type ServiceDeployStatus struct {
	Name                 string
	BuildStatus          source.BuildStatus
	CurrentDeployVersion int
	NextDeployVersion    int
}

// Deploy ...
func (h *LocalHandler) Deploy(ctx context.Context, p *DeployParams) (*DeployResp, error) {
	source, err := h.Source.Get(ctx, p.ProjectName, p.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	if !source.ChecksComplete {
		return &DeployResp{
			Source:   source,
			Services: mapServices(source, nil),
		}, nil
	}

	// no confirm yet, so don't do the deploy
	if p.Confirm == nil {
		dest, err := h.Dest.Get(ctx, p.ProjectName)
		if err != nil {
			return nil, fmt.Errorf("failed to get destination: %w", err)
		}

		return &DeployResp{
			Source:   source,
			Services: mapServices(source, dest),
		}, nil
	}

	// Check the confirm git hash lines up
	if p.Confirm.CommitHash != source.Ref.CommitHash {
		return nil, fmt.Errorf("source git hash %s does not match confirm git hash %s", source.Ref.CommitHash, p.Confirm.CommitHash)
	}

	depreq := &destination.DeployParams{
		ProjectName: p.ProjectName,
		Services:    []*destination.ServiceDeployParams{},
	}

	for _, creq := range p.Confirm.Requests {
		svc := source.Services.Lookup(creq.ServiceName)
		if svc == nil {
			return nil, fmt.Errorf("service %s not found in source", creq.ServiceName)
		}

		slug, err := SlugifyServiceName(creq.ServiceName)
		if err != nil {
			return nil, fmt.Errorf("failed to slugify service name %s: %w", creq.ServiceName, err)
		}

		shorthash := source.Ref.CommitHash.Short()

		depreq.Services = append(depreq.Services, &destination.ServiceDeployParams{
			ServiceConfig: svc.Service,
			Name:          creq.ServiceName,
			SlugName:      fmt.Sprintf("%s-v%d", slug, creq.DeployVersion),
			Version:       creq.DeployVersion,
			DeployImage:   fmt.Sprintf("%s:%s", path.Join(source.Project.RegistryPrefix, creq.ServiceName), shorthash),
			Namespace:     source.Project.Gitops.Namespace,
		})
	}

	if _, err := h.Dest.Deploy(ctx, depreq); err != nil {
		return nil, fmt.Errorf("failed to deploy destination: %w", err)
	}

	return &DeployResp{
		Source:   source,
		Complete: true,
	}, nil
}

func mapServices(source *source.Source, dest *destination.Destination) []*ServiceDeployStatus {
	services := []*ServiceDeployStatus{}

	for _, s := range source.Services {
		s2 := &ServiceDeployStatus{
			Name:        s.Name,
			BuildStatus: s.BuildStatus,
		}

		if dest != nil {
			s2.CurrentDeployVersion = dest.Services.CurrentDeployVersionFor(s.Name)
			s2.NextDeployVersion = dest.Services.NextDeployVersionFor(s.Name)
		}

		services = append(services, s2)
	}

	return services
}
