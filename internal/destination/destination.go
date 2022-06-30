package destination

import "github.com/cygnetdigital/shipper"

// Destination encapsulates the current state of a destination
type Destination struct {
	ProjectName string
	Services    Services
}

// NextDeployVersion returns the next deploy version for the given service
func (d *Destination) NextDeployVersion(project, service string) (string, error) {
	svc := d.Services.LookupByProjectAndName(project, service)
	if svc == nil {
		return firstVersionString, nil
	}

	return svc.NextDeployVersion()
}

// DeployParams are required to create a new deployment in the destination
type DeployParams struct {
	ProjectName string
	Services    []*ServiceDeployParams
}

// ServiceDeployParams are the required params to deploy a service
type ServiceDeployParams struct {
	Config   *shipper.Service
	Version  string
	ImageTag string
}

// DeployResp is the response from a deploy
type DeployResp struct {
	Hash string
}

// ReleaseParams ...
type ReleaseParams struct {
	Project string
	Service string
	Version string
}

// ReleaseResp ...
type ReleaseResp struct {
	Hash string
}

// RemoveParams ...
type RemoveParams struct {
	Project string
	Service string
	Version string
}

// RemoveResp ...
type RemoveResp struct {
	Hash string
}
