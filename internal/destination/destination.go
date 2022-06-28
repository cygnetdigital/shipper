package destination

import "github.com/cygnetdigital/shipper"

// Destination encapsulates the current state of a destination
type Destination struct {
	ProjectName string
	Services    Services
}

// DeployParams are required to create a new deployment in the destination
type DeployParams struct {
	ProjectName string
	Services    []*ServiceDeployParams
}

// ServiceDeployParams are the required params to deploy a service
type ServiceDeployParams struct {
	ServiceConfig *shipper.Service
	Name          string
	Version       int
	SlugName      string
	DeployImage   string
	Namespace     string
}

// DeployResp is the response from a deploy
type DeployResp struct {
	Hash string
}
