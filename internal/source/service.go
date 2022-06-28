package source

import (
	"fmt"
	"time"

	"github.com/cygnetdigital/shipper"
)

// Services is a list of services
type Services []*Service

// Lookup a service by name
func (s Services) Lookup(name string) *Service {
	for _, svc := range s {
		if svc.Name == name {
			return svc
		}
	}

	return nil
}

// Service wraps the core service config with the build status
type Service struct {
	*shipper.Service
	BuildStatus BuildStatus
}

// BuildStatus represents a number of build states
type BuildStatus interface {
	String() string
}

// BuildStatusQueued ...
type BuildStatusQueued struct{}

func (b BuildStatusQueued) String() string {
	return "➡️ queued"
}

// BuildStatusRunning ...
type BuildStatusRunning struct {
	StartedAt time.Time
}

func (b BuildStatusRunning) String() string {
	return "🏃 running"
}

// BuildStatusComplete ...
type BuildStatusComplete struct {
	StartedAt  time.Time
	FinishedAt time.Time
}

func (b BuildStatusComplete) String() string {
	return "✅ complete"
}

// BuildStatusFailed ...
type BuildStatusFailed struct {
	Reason string
}

func (b BuildStatusFailed) String() string {
	return fmt.Sprintf("❌  failed - %s", b.Reason)
}
