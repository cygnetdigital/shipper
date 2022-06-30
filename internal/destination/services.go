package destination

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Service ...
type Service struct {
	// Name of the service
	Name string

	// Name of the project
	Project string

	// Bundles of manifest files that this service has
	Deploys []*Deploy

	// Release info
	Release *Release

	// Currently released version
	CurrentReleaseVersion string
}

func (s *Service) process() error {
	// numerically sort deploys in ascending order
	sort.Slice(s.Deploys, func(i, j int) bool {
		ii, _ := strconv.Atoi(strings.TrimPrefix(s.Deploys[i].Version, "v"))
		jj, _ := strconv.Atoi(strings.TrimPrefix(s.Deploys[j].Version, "v"))

		return ii < jj
	})

	return nil
}

// HasVersion returns true if this version exists
func (s *Service) HasVersion(version string) bool {
	for _, d := range s.Deploys {
		if d.Version == version {
			return true
		}
	}

	return false
}

// Deploy ...
type Deploy struct {
	Project   string
	Name      string
	Version   string
	Manifests []*Manifest
}

// Release ...
type Release struct {
	Project   string
	Name      string
	Manifests []*Manifest
}

// LoadServices produces services by parsing kubernetes manifests and
// mapping over annotations according to the desired spec.
func LoadServices(rootDir string) (Services, error) {
	manifests := []*Manifest{}

	filepaths, err := glob(rootDir)
	if err != nil {
		return nil, err
	}

	for _, fp := range filepaths {
		mfs, err := openAndParseManifest(fp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file %s: %w", fp, err)
		}

		manifests = append(manifests, mfs...)
	}

	svcs := Services{}

	for _, mf := range manifests {
		switch mf.Annotations["shipper/bundle"] {
		case "deploy":
			svcs = svcs.appendDeploy(mf)

		case "release":
			svcs = svcs.appendRelease(mf)

		default:
			continue
		}
	}

	if err := svcs.process(); err != nil {
		return nil, err
	}

	return svcs, nil
}

// Services ...
type Services []*Service

// FilterByProject name
func (s Services) FilterByProject(proj string) (out Services) {
	for _, service := range s {
		if service.Project == proj {
			out = append(out, service)
		}
	}

	return out
}

// Lookup a service by name
func (s Services) Lookup(name string) *Service {
	for _, service := range s {
		if service.Name == name {
			return service
		}
	}

	return nil
}

// LookupByProjectAndName ...
func (s Services) LookupByProjectAndName(proj, name string) *Service {
	for _, service := range s {
		if service.Project == proj && service.Name == name {
			return service
		}
	}

	return nil
}

// appendDeploy will build a service from a manifest, or append itself
// to a deploy bundle within an already existing service
func (s Services) appendDeploy(mf *Manifest) Services {
	dep := &Deploy{
		Project:   mf.Annotations["shipper/project"],
		Name:      mf.Annotations["shipper/service-name"],
		Version:   mf.Annotations["shipper/service-version"],
		Manifests: []*Manifest{mf},
	}

	// skip if no project/service-name
	if dep.Project == "" || dep.Name == "" || dep.Version == "" {
		return s
	}

	// look for a matching service
	for _, svc := range s {
		if svc.Project == dep.Project && svc.Name == dep.Name {
			// on a matching service, match against a deploy version
			for _, db := range svc.Deploys {
				if db.Version == dep.Version {
					db.Manifests = append(db.Manifests, mf)

					return s
				}
			}

			// reaching here means this is the first manifest for this deploy
			svc.Deploys = append(svc.Deploys, dep)

			return s
		}
	}

	return append(s, &Service{
		Project: dep.Project,
		Name:    dep.Name,
		Deploys: []*Deploy{dep},
	})
}

// appendRelease will build a service from a manifest, or append itself to
// an already existing service
func (s Services) appendRelease(mf *Manifest) Services {
	rel := &Release{
		Project:   mf.Annotations["shipper/project"],
		Name:      mf.Annotations["shipper/service-name"],
		Manifests: []*Manifest{mf},
	}

	// skip if no project/service-name
	if rel.Project == "" || rel.Name == "" {
		return s
	}

	release := mf.Annotations["shipper/current-release"]

	for _, svc := range s {
		if svc.Project == rel.Project && svc.Name == rel.Name {
			if svc.Release != nil {
				svc.Release.Manifests = append(svc.Release.Manifests, mf)
			} else {
				svc.Release = rel
			}

			if release != "" {
				svc.CurrentReleaseVersion = release
			}

			return s
		}
	}

	return append(s, &Service{
		Project:               rel.Project,
		Name:                  rel.Name,
		Release:               rel,
		CurrentReleaseVersion: release,
	})
}

// process the various fields across services based on manifest
func (s Services) process() error {
	for _, svc := range s {
		if err := svc.process(); err != nil {
			return err
		}
	}

	return nil
}
