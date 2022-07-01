package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cygnetdigital/shipper/internal/conf"
	"github.com/cygnetdigital/shipper/internal/destination"
)

// Deploy to the destination
func (s *Github) Deploy(ctx context.Context, p *destination.DeployParams) (*destination.DeployResp, error) {
	if p.ProjectName != s.projName {
		return nil, fmt.Errorf("project %s not supported", p.ProjectName)
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

	templateRoot := path.Join(rootPath, s.templatePath)
	manifestRoot := path.Join(rootPath, s.bundlePath)

	for _, sp := range p.Services {
		svc := svcs.LookupByProjectAndName(p.ProjectName, sp.Config.Name)

		if svc != nil && svc.HasVersion(sp.Version) {
			return nil, fmt.Errorf("version already exists")
		}

		slug, err := destination.SlugifyServiceName(sp.Config.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to slugify service name: %w", err)
		}

		dv, err := setupDeployVariables(sp.Config.Deploy.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup deploy variables: %w", err)
		}

		args := &destination.DeployContext{
			Project:         p.ProjectName,
			Name:            sp.Config.Name,
			Version:         sp.Version,
			SlugName:        slug,
			SlugNameVersion: fmt.Sprintf("%s-%s", slug, sp.Version),
			DeployImage:     fmt.Sprintf("%s:%s", path.Join(s.registry, sp.Config.Name), sp.ImageTag),
			DeployVariables: dv,
			SecretMounts:    sp.Config.Deploy.SecretMounts,
			Namespace:       s.namespace,
		}

		template := sp.Config.Deploy.Template

		if err := destination.WriteDeployBundle(templateRoot, manifestRoot, template, args); err != nil {
			return nil, fmt.Errorf("failed to write bundle: %w", err)
		}
	}

	msg := fmt.Sprintf("Deploying %s", p.ProjectName)
	if len(p.Services) == 1 {
		msg = fmt.Sprintf("Deploying %s/%s", p.ProjectName, p.Services[0].Config.Name)
	}

	hash, err := s.commit(msg, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &destination.DeployResp{Hash: hash}, nil
}

func setupDeployVariables(configs []*conf.ServiceConfigItem) (map[string]string, error) {
	out := map[string]string{}

	for _, c := range configs {
		v, err := json.Marshal(c.Values)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("SHIPPER_%s", strings.ToUpper(c.Name))
		out[key] = string(v)
	}

	return out, nil
}
