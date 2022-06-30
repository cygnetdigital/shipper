package cli

import (
	"fmt"
	"os"

	"github.com/cygnetdigital/shipper"
	"github.com/cygnetdigital/shipper/internal/cliutil"
	"github.com/cygnetdigital/shipper/internal/destination/github"
	"github.com/cygnetdigital/shipper/internal/source"
	"github.com/cygnetdigital/shipper/pkg/handler"
	"github.com/urfave/cli/v2"
)

// Release command
var Release = &cli.Command{
	Name:        "release",
	Usage:       "generate release manifests and push them to a gitops repository",
	Description: "e.g. `shipper release service.foo`",
	ArgsUsage:   "[service]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "github-token",
			EnvVars: []string{
				"SHIPPER_GITHUB_TOKEN",
			},
		},
		&cli.StringFlag{
			Name:  "version",
			Usage: "version to release instead of the latest",
		},
	},
	Action: func(c *cli.Context) error {
		ght := c.String("github-token")
		if ght == "" {
			return fmt.Errorf("github-token is required")
		}

		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working dir: %w", err)
		}

		proj, err := shipper.LoadProject(pwd)
		if err != nil {
			return fmt.Errorf("failed to get project context: %w", err)
		}

		hand := &handler.LocalHandler{
			Source: source.NewGithub(proj, ght),
			Dest:   github.NewGithub(proj, ght),
		}

		v := c.String("version")

		rp := &handler.ReleaseParams{
			Project: proj.Name,
			Service: c.Args().First(),
			Version: v,
		}

		rres, err := hand.Release(c.Context, rp)
		if err != nil {
			return fmt.Errorf("failed to release: %w", err)
		}

		if rres.Done {
			if v == "" {
				fmt.Printf("%s is already at the latest version (%s)\n", rres.Service, rres.Version)
			} else {
				fmt.Printf("%s is already at %s\n", rres.Service, rres.Version)
			}

			return nil
		}

		resp := cliutil.StringPrompt(fmt.Sprintf("Release %s → %s ?", rres.Service, rres.Version))
		if resp != "YES" {
			return fmt.Errorf("aborted: only YES is accepted")
		}

		rp.Confirm = true

		rres2, err := hand.Release(c.Context, rp)
		if err != nil {
			return fmt.Errorf("failed to release: %w", err)
		}

		if rres2.Done {
			fmt.Println("done")
		}

		return nil
	},
}
