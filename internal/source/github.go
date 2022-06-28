package source

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cygnetdigital/shipper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// Github allows a github source to be used
type Github struct {
	name        string
	repo        string
	ensureClean bool

	// gh *github.Client
	gh *GithubHelper

	gitAuth *http.BasicAuth
}

// NewGithub sets up a new github source
func NewGithub(proj *shipper.Project, accessToken string) *Github {
	if proj.Repo == "" {
		panic("project repo is required")
	}

	if !strings.Contains(proj.Repo, "github.com") {
		panic("project repo must be a github repo")
	}

	owner, repo, found := strings.Cut(strings.TrimPrefix(proj.Repo, "github.com/"), "/")
	if !found {
		panic("project repo must be of form 'github.com/org/repo'")
	}

	return &Github{
		name:        proj.Name,
		repo:        proj.Repo,
		ensureClean: false,
		gh: &GithubHelper{
			client: github.NewClient(
				oauth2.NewClient(
					context.Background(),
					oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
				),
			),
			owner: owner,
			repo:  repo,
		},
		gitAuth: &http.BasicAuth{Username: "username", Password: accessToken},
	}
}

// Get source from github
func (s *Github) Get(ctx context.Context, projectName string, ref string) (*Source, error) {
	if projectName != s.name {
		return nil, fmt.Errorf("project '%s' not setup", projectName)
	}

	// resolve the given ref (could be a branch or PR number)
	resolvedRef, err := s.gh.Resolve(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup ref: %w", err)
	}

	out := &Source{
		ProjectName: projectName,
		Ref:         resolvedRef,
	}

	// if there is no commit hash, the PR is probably not merged, so we can't
	// do anything.
	if resolvedRef.CommitHash == GitHash("") {
		return out, nil
	}

	// clone the repo and checkout the commit hash
	repoPath, err := s.clone(resolvedRef.CommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repo: %w", err)
	}

	// cleanup when ready
	//nolint:errcheck
	defer os.RemoveAll(repoPath)

	// load the project configuration at this commit
	proj, err := shipper.LoadProject(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get project context: %w", err)
	}

	// get the github checks for this commit
	checks, err := s.gh.getWorkflowChecks(ctx, resolvedRef.CommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get checks: %w", err)
	}

	out.ChecksRunning = checks.Run.GetStatus() != "completed"
	out.ChecksComplete = checks.Run.GetConclusion() == "success"
	out.Project = proj
	out.Services, err = buildServices(proj, checks.Jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to build services: %w", err)
	}

	return out, nil
}

func (s *Github) clone(hash GitHash) (string, error) {
	temp, err := os.MkdirTemp("", "dir")
	if err != nil {
		return "", err
	}

	uri, err := url.Parse(s.repo)
	if err != nil {
		return "", err
	}

	if uri.Scheme == "" {
		uri.Scheme = "https"
	}

	repo, err := git.PlainClone(temp, false, &git.CloneOptions{
		URL:           uri.String(),
		Depth:         20,
		Auth:          s.gitAuth,
		ReferenceName: plumbing.ReferenceName("refs/heads/main"),
		SingleBranch:  true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}

	commit, err := repo.CommitObject(plumbing.NewHash(string(hash)))
	if err != nil {
		return temp, fmt.Errorf("failed to get commit: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return temp, fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := wt.Checkout(&git.CheckoutOptions{Hash: commit.Hash}); err != nil {
		return temp, fmt.Errorf("failed to checkout commit: %w", err)
	}

	return temp, nil
}
