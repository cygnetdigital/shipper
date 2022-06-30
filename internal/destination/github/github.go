package github

import (
	"fmt"
	"net/url"
	"os"

	"github.com/cygnetdigital/shipper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Github destination is capable of deploying manifests to github
type Github struct {
	projName     string
	repo         string
	templatePath string
	bundlePath   string
	registry     string
	namespace    string
	auth         *http.BasicAuth
}

// NewGithub sets up a git destination
func NewGithub(proj *shipper.Project, ghToken string) *Github {
	return &Github{
		projName:     proj.Name,
		repo:         proj.Gitops.Repo,
		templatePath: proj.Gitops.TemplatePath,
		bundlePath:   proj.Gitops.ManifestPath,
		registry:     proj.RegistryPrefix,
		namespace:    proj.Gitops.Namespace,
		auth:         &http.BasicAuth{Username: "username", Password: ghToken},
	}
}

// clone the repo
func (s *Github) clone() (*git.Repository, string, error) {
	temp, err := os.MkdirTemp("", "dir")
	if err != nil {
		return nil, "", err
	}

	uri, err := url.Parse(s.repo)
	if err != nil {
		return nil, "", err
	}

	if uri.Scheme == "" {
		uri.Scheme = "https"
	}

	repo, err := git.PlainClone(temp, false, &git.CloneOptions{
		URL:   uri.String(),
		Depth: 1,
		Auth:  s.auth,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to clone repo: %w", err)
	}

	return repo, temp, nil
}

func (s *Github) commit(msg string, repo *git.Repository) (string, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := wt.AddGlob(s.bundlePath + "/*"); err != nil {
		return "", fmt.Errorf("failed to add files to worktree: %w", err)
	}

	hash, err := wt.Commit(msg, &git.CommitOptions{All: true})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	if err := repo.Push(&git.PushOptions{RemoteName: "origin", Auth: s.auth}); err != nil {
		return "", fmt.Errorf("failed to push: %w", err)
	}

	return hash.String(), nil
}
