package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/cygnetdigital/shipper"
	"github.com/cygnetdigital/shipper/internal/cliutil"
	"github.com/cygnetdigital/shipper/internal/destination/github"
	"github.com/cygnetdigital/shipper/internal/source"
	"github.com/cygnetdigital/shipper/pkg/handler"
	"github.com/urfave/cli/v2"
)

// Deploy command
var Deploy = &cli.Command{
	Name:        "deploy",
	Usage:       "generate kubernetes manifests and push them to a gitops repository",
	Description: "e.g. `shipper deploy 123` or `shipper deploy feature/foo`",
	ArgsUsage:   "[ref]",
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

		dp := &handler.DeployParams{
			ProjectName: proj.Name,
			Ref:         c.Args().First(),
		}

		var dres *handler.DeployResp

		printer := cliutil.NewDeployPrinter()

		for {
			var err error
			dres, err = hand.Deploy(c.Context, dp)
			if err != nil {
				return fmt.Errorf("failed to deploy: %w", err)
			}

			if printer.Print(dres) {
				time.Sleep(time.Second)

				continue
			}

			break
		}

		printer.Stop()

		if dres.Source.Ref.CommitHash == "" {
			return nil
		}

		if len(dres.Services) == 0 {
			return nil
		}

		resp := cliutil.StringPrompt("Deploy to production?")
		if resp != "YES" {
			return fmt.Errorf("aborted: only YES is accepted")
		}

		dres2, err := hand.Deploy(c.Context, &handler.DeployParams{
			ProjectName: proj.Name,
			Ref:         c.Args().First(),
			Confirm: &handler.ConfirmDeployParams{
				CommitHash: dres.Source.Ref.CommitHash,
				Requests:   buildRequestsForSerivcse(dres.Services),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to deploy: %w", err)
		}

		if !dres2.Complete {
			return fmt.Errorf("deployment failed")
		}

		fmt.Printf("\nDeployment complete\n")

		return nil
	},
}

// BuildRequests ...
func buildRequestsForSerivcse(svcs []*handler.ServiceDeployStatus) (out []*handler.ServiceDeployRequest) {
	for _, s := range svcs {
		out = append(out, &handler.ServiceDeployRequest{
			ServiceName:   s.Name,
			DeployVersion: s.NextDeployVersion,
		})
	}

	return out
}
