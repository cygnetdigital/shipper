package handler

import (
	"context"

	"github.com/cygnetdigital/shipper/internal/destination"
	"github.com/cygnetdigital/shipper/internal/source"
)

// LocalHandler coordinates the various shipper commands with state from sources
// and destinations. This handler would be used directly within the cmds.
type LocalHandler struct {
	Source SourceGetter
	Dest   Destination
}

// SourceGetter allows us to get source for a project/ref
type SourceGetter interface {
	Get(ctx context.Context, project string, ref string) (*source.Source, error)
}

// Destination ...
type Destination interface {
	Get(ctx context.Context, project string) (*destination.Destination, error)
	Deploy(ctx context.Context, p *destination.DeployParams) (*destination.DeployResp, error)
}
