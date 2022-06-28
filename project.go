package shipper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cygnetdigital/shipper/internal/conf"
	"gopkg.in/yaml.v3"
)

// Project contains information about the project
type Project struct {
	*conf.Project
	RootDir string

	// Services part of this project
	Services Services
}

var projectContextFile = "shipper.project.yaml"

// LoadProject configuration at pwd
func LoadProject(pwd string) (*Project, error) {
	paths := getParentPaths(pwd)

	for _, p := range paths {
		confPath := filepath.Join(p, projectContextFile)

		bts, err := ioutil.ReadFile(confPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to read config %s: %w", confPath, err)
		}

		var proj conf.Project
		if err := yaml.Unmarshal(bts, &proj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conf %s: %w", confPath, err)
		}

		ctx := &Project{
			Project: &proj,
			RootDir: p,
		}

		expandedPaths, err := expandWildcardPaths(p, proj.Paths)
		if err != nil {
			return nil, fmt.Errorf("failed to expand wildcard for %s: %w", confPath, err)
		}

		for _, sp := range expandedPaths {
			svcConfPath := filepath.Join(sp, "shipper.yaml")

			bts, err := ioutil.ReadFile(svcConfPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return nil, fmt.Errorf("failed to read config %s: %w", svcConfPath, err)
			}

			var svc conf.Service
			if err := yaml.Unmarshal(bts, &svc); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conf %s: %w", svcConfPath, err)
			}

			pathRelToProject := strings.TrimPrefix(strings.TrimPrefix(sp, p), "/")

			ctx.Services = append(ctx.Services, &Service{Service: &svc, RootDir: pathRelToProject})
		}

		return ctx, nil
	}

	return nil, fmt.Errorf("%s not found in this or any parent directory", projectContextFile)
}
