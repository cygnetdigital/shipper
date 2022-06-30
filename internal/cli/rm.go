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

// Remove command
var Remove = &cli.Command{
	Name:        "remove",
	Aliases:     []string{"rm"},
	Usage:       "remove a deployment from a gitops repository",
	Description: "e.g. `shipper rm service.foo v1`",
	ArgsUsage:   "[service] [version]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "github-token",
			EnvVars: []string{
				"SHIPPER_GITHUB_TOKEN",
			},
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

		rp := &handler.RemoveParams{
			Project: proj.Name,
			Service: c.Args().Get(0),
			Version: c.Args().Get(1),
		}

		rres, err := hand.Remove(c.Context, rp)
		if err != nil {
			return fmt.Errorf("failed to remove: %w", err)
		}

		resp := cliutil.StringPrompt(fmt.Sprintf("Remove %s @ %s ?", rres.Service, rres.Version))
		if resp != "YES" {
			return fmt.Errorf("aborted: only YES is accepted")
		}

		rp.Confirm = true

		rres2, err := hand.Remove(c.Context, rp)
		if err != nil {
			return fmt.Errorf("failed to remove: %w", err)
		}

		if rres2.Done {
			fmt.Println("done")
		}

		return nil
	},
}
