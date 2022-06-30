package source

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cygnetdigital/shipper"
	"github.com/google/go-github/v45/github"
)

// GithubHelper ...
type GithubHelper struct {
	client *github.Client
	owner  string
	repo   string
}

// Resolve a ref using the github API. Currently supporting a ref which is a
// PullRequest number or branch name.
func (g *GithubHelper) Resolve(ctx context.Context, ref string) (*Ref, error) {
	if ref == "main" {
		branch, _, err := g.client.Repositories.GetBranch(ctx, g.owner, g.repo, "main", false)
		if err != nil {
			return nil, fmt.Errorf("failed to get main branch: %w", err)
		}

		return &Ref{
			GivenRef:           "main",
			CommitHash:         GitHash(branch.GetCommit().GetSHA()),
			CommitedByUsername: branch.GetCommit().GetAuthor().GetLogin(),
		}, nil
	}

	// resolve a integer (presumed to be a PR number)
	if n, err := strconv.Atoi(ref); err == nil {
		pr, _, err := g.client.PullRequests.Get(ctx, g.owner, g.repo, n)
		if err != nil {
			return nil, fmt.Errorf("failed to get pull request: %w", err)
		}

		return buildRefForPR(pr), nil
	}

	prs, _, err := g.client.PullRequests.List(ctx, g.owner, g.repo, &github.PullRequestListOptions{
		Head: fmt.Sprintf("%s:%s", g.owner, ref),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	if len(prs) == 0 {
		return nil, fmt.Errorf("no pull requests found for branch '%s'", ref)
	}

	return buildRefForPR(prs[0]), nil
}

func buildRefForPR(pull *github.PullRequest) *Ref {
	pr := &PullRequest{
		Title:      pull.GetTitle(),
		Number:     pull.GetNumber(),
		URL:        pull.GetHTMLURL(),
		HeadCommit: *buildCommitForPRBranch(pull.GetHead()),
		BaseCommit: *buildCommitForPRBranch(pull.GetBase()),
		Merged:     pull.GetMerged(),
		MergedAt:   pull.GetMergedAt(),
	}

	if pr.Merged {
		pr.MergeCommitHash = GitHash(pull.GetMergeCommitSHA())
	}

	if mb := pull.GetMergedBy(); mb != nil {
		pr.MergedByUsername = mb.GetLogin()
	}

	return &Ref{
		GivenRef:    fmt.Sprintf("Pull Request #%d", pull.GetID()),
		PullRequest: pr,
		CommitHash:  pr.MergeCommitHash,
	}
}

func buildCommitForPRBranch(branch *github.PullRequestBranch) *GithubCommit {
	return &GithubCommit{
		Ref:      branch.GetRef(),
		Hash:     GitHash(branch.GetSHA()),
		Username: branch.GetUser().GetLogin(),
	}
}

func (g *GithubHelper) getWorkflowChecks(ctx context.Context, hash GitHash) (*workflowCheck, error) {
	suites, _, err := g.client.Checks.ListCheckSuitesForRef(ctx, g.owner, g.repo, string(hash), &github.ListCheckSuiteOptions{
		ListOptions: github.ListOptions{PerPage: 1000},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get check suites: %w", err)
	}

	if len(suites.CheckSuites) == 0 {
		return nil, fmt.Errorf("no github check suites found for git hash %s", hash)
	}

	u := fmt.Sprintf("repos/%s/%s/actions/runs?check_suite_id=%d", g.owner, g.repo, suites.CheckSuites[0].GetID())

	req, err := g.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	runs := new(github.WorkflowRuns)
	if _, err := g.client.Do(ctx, req, &runs); err != nil {
		return nil, err
	}

	if len(runs.WorkflowRuns) == 0 {
		return nil, fmt.Errorf("no Github workflows ran for %s", hash)
	}

	run := runs.WorkflowRuns[0]

	jobs, _, err := g.client.Actions.ListWorkflowJobs(ctx, g.owner, g.repo, run.GetID(), &github.ListWorkflowJobsOptions{
		ListOptions: github.ListOptions{PerPage: 1000},
	})
	if err != nil {
		return nil, err
	}

	return &workflowCheck{
		Run:  run,
		Jobs: jobs.Jobs,
	}, nil
}

type workflowCheck struct {
	Run  *github.WorkflowRun
	Jobs []*github.WorkflowJob
}

func buildServices(proj *shipper.Project, jobs []*github.WorkflowJob) ([]*Service, error) {
	svcs := make([]*Service, 0, len(proj.Services))

	for _, svc := range proj.Services {
		s := &Service{
			Service: svc,
		}

		var seen bool

		// go through each check and find the ones that match the service name
		for _, job := range jobs {
			if extractSvcName(job.GetName()) != s.Name {
				continue
			}

			if seen {
				return nil, fmt.Errorf("duplicate workflow jobs found for %s, and shipper only ever expected one per service", s.Name)
			}

			seen = true
			s.BuildStatus = statusForJob(job)
		}

		if seen {
			svcs = append(svcs, s)
		}
	}

	return svcs, nil
}

var re = regexp.MustCompile(`(service\.[a-zA-Z0-9\-\.]+)`)

func extractSvcName(checkName string) string {
	submatch := re.FindStringSubmatch(checkName)
	if len(submatch) == 0 {
		return ""
	}

	return submatch[0]
}

//nolint
func statusForJob(job *github.WorkflowJob) BuildStatus {
	switch job.GetStatus() {
	case "queued":
		return &BuildStatusQueued{}

	case "in_progress":
		return &BuildStatusRunning{
			StartedAt: job.GetStartedAt().Time,
		}

	case "completed":
		c := job.GetConclusion()
		switch c {
		case "success":
			return &BuildStatusComplete{
				StartedAt:  job.GetStartedAt().Time,
				FinishedAt: job.GetCompletedAt().Time,
			}

		default:
			return &BuildStatusFailed{
				Reason: fmt.Sprintf("job conclusion of '%s'", c),
			}
		}

	default:
		return &BuildStatusFailed{
			Reason: "unknown job status",
		}
	}
}
