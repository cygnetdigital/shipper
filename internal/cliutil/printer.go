package cliutil

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cygnetdigital/shipper/internal/source"
	"github.com/cygnetdigital/shipper/pkg/handler"
	"github.com/gosuri/uilive"
)

// DeployPrinter can print handler responses to terminal with live updating
type DeployPrinter struct {
	writer     *uilive.Writer
	firstPrint bool
	started    bool
}

// NewDeployPrinter sets up a new deploy printer
func NewDeployPrinter() *DeployPrinter {
	return &DeployPrinter{
		writer:     uilive.New(),
		firstPrint: true,
	}
}

// Print returns true if another print attempt should be made
func (dp *DeployPrinter) Print(r *handler.DeployResp) bool {
	if dp.firstPrint {
		dp.firstPrint = false

		ref := r.Source.Ref
		fmt.Printf("🔍  Resolving ref as %s\n", ref.GivenRef)

		if ref.CommitHash == "" {
			fmt.Println("❌  Commit not found for ref")

			return false
		}

		pr := ref.PullRequest
		if pr != nil {
			fmt.Printf("💻  #%d: %s\n", pr.Number, pr.Title)

			if !pr.Merged {
				fmt.Printf("❌  PR not merged, cannot continue deploy\n")

				return false
			}

			fmt.Printf("🔀  Merged into %s by %s → %s\n", pr.BaseCommit.Ref, pr.MergedByUsername, pr.MergeCommitHash)
		} else {
			fmt.Printf("🔀  Commit %s found by %s\n", ref.CommitHash, ref.CommitedByUsername)
		}

		// if checks are running, lets start the writer
		if r.Source.ChecksRunning {
			dp.start()

			return true
		}

		dp.printServices(os.Stdout, r)

		return false
	}

	dp.printServices(dp.writer, r)

	// return true if printing should happen again
	return r.Source.ChecksRunning
}

func (dp *DeployPrinter) printServices(w io.Writer, resp *handler.DeployResp) {
	if resp.Source.ChecksRunning {
		fmt.Fprintf(w, "\n🏗️   Waiting for checks to complete\n")
	} else {
		fmt.Fprintf(w, "\n✅  Ready to deploy:\n")
	}

	if len(resp.Services) == 0 {
		return
	}

	longestName := 0
	for _, s := range resp.Services {
		if len(s.Name) > longestName {
			longestName = len(s.Name)
		}
	}

	for _, svc := range resp.Services {
		nameWithPadding := fmt.Sprintf("%s%s", svc.Name, strings.Repeat(" ", longestName+5-len(svc.Name)))

		switch svc.BuildStatus.(type) {
		case *source.BuildStatusComplete:
			if svc.NextDeployVersion != "" {
				fmt.Fprintf(w, "   👉  %s → %s\n", nameWithPadding, svc.NextDeployVersion)
			} else {
				fmt.Fprintf(w, "   👉  %s %s\n", nameWithPadding, svc.BuildStatus.String())
			}

		default:
			fmt.Fprintf(w, "   👉  %s %s\n", nameWithPadding, svc.BuildStatus.String())
		}
	}

	fmt.Fprintf(w, "\n")
}

func (dp *DeployPrinter) start() {
	dp.started = true
	dp.writer.Start()
}

// Stop the live writer if it started
func (dp *DeployPrinter) Stop() {
	if dp.started {
		dp.writer.Stop()
	}
}
