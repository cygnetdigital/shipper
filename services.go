package shipper

import (
	"strings"

	"github.com/cygnetdigital/shipper/internal/conf"
)

// Service represents a service parsed config
type Service struct {
	*conf.Service
	RootDir string
}

// Services are the services in the project
type Services []*Service

// Lookup a service by name
func (s Services) Lookup(ref string) Services {
	out := Services{}

	for _, name := range strings.Split(ref, ",") {
		for _, svc := range s {
			if svc.Name == strings.TrimSpace(name) {
				out = append(out, svc)
			}
		}
	}

	return out
}

// Names of all services
func (s Services) Names() (out []string) {
	for _, svc := range s {
		out = append(out, svc.Name)
	}

	return
}
