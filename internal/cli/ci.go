package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cygnetdigital/shipper"
	"github.com/urfave/cli/v2"
)

// CI command
var CI = &cli.Command{
	Name:  "ci",
	Usage: "ci related helper commands",
	Flags: []cli.Flag{},
	Subcommands: []*cli.Command{
		{
			Name:        "list-services",
			Usage:       "list services to build",
			Description: "e.g. `shipper ci list-services`",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "changed-file-filter",
					Usage: "filter the services to build based on the changed files",
				},
			},
			Action: func(c *cli.Context) error {
				pwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working dir: %w", err)
				}

				proj, err := shipper.LoadProject(pwd)
				if err != nil {
					return fmt.Errorf("failed to get project context: %w", err)
				}

				type service struct {
					Name       string `json:"name"`
					Dockerfile string `json:"dockerfile"`
					Path       string `json:"path"`
				}

				svcs := []*service{}
				filter := newChangedFileFilter(c.String("changed-file-filter"))

				for _, svc := range proj.Services {
					if !filter.Include(svc) {
						continue
					}

					svcs = append(svcs, &service{
						Name:       svc.Name,
						Dockerfile: svc.Build.Dockerfile,
						Path:       svc.RootDir,
					})
				}

				if err := json.NewEncoder(os.Stdout).Encode(svcs); err != nil {
					return fmt.Errorf("failed to encode json: %w", err)
				}

				return nil
			},
		},
	},
}

type changedFileFilter struct {
	active bool
	files  []string
}

func newChangedFileFilter(files string) *changedFileFilter {
	return &changedFileFilter{
		files:  strings.Split(files, " "),
		active: files != "",
	}
}

// Include returns true if the service should be built
func (f *changedFileFilter) Include(svc *shipper.Service) bool {
	if !f.active {
		return true
	}

	for _, file := range f.files {
		if strings.HasPrefix(file, svc.RootDir) {
			return true
		}
	}

	return false
}
