package github

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/cygnetdigital/shipper"
	"github.com/cygnetdigital/shipper/internal/destination"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Github destination is capable of deploying manifests to github
type Github struct {
	projName     string
	repo         string
	templatePath string
	bundlePath   string
	auth         *http.BasicAuth
}

// NewGithub sets up a git destination
func NewGithub(proj *shipper.Project, ghToken string) *Github {
	return &Github{
		projName:     proj.Name,
		repo:         proj.Gitops.Repo,
		templatePath: proj.Gitops.TemplatePath,
		bundlePath:   proj.Gitops.ManifestPath,
		auth:         &http.BasicAuth{Username: "username", Password: ghToken},
	}
}

// Get the destination state
func (s *Github) Get(ctx context.Context, projectName string) (*destination.Destination, error) {
	if projectName != s.projName {
		return nil, fmt.Errorf("project %s not supported", projectName)
	}

	_, path, err := s.clone()
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer os.RemoveAll(path)

	services, err := destination.ParseStateFiles(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipper.state.yaml files: %w", err)
	}

	return &destination.Destination{
		ProjectName: projectName,
		Services:    services.FilterByProject(s.projName),
	}, nil
}

// Deploy to the destination
func (s *Github) Deploy(ctx context.Context, p *destination.DeployParams) (*destination.DeployResp, error) {
	if p.ProjectName != s.projName {
		return nil, fmt.Errorf("project %s not supported", p.ProjectName)
	}

	repo, rootPath, err := s.clone()
	if err != nil {
		return nil, err
	}

	svcs, err := destination.ParseStateFiles(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipper.state.yaml files: %w", err)
	}

	svcs = svcs.FilterByProject(p.ProjectName)

	//nolint:errcheck
	defer os.RemoveAll(rootPath)

	templateRoot := path.Join(rootPath, s.templatePath)
	manifestRoot := path.Join(rootPath, s.bundlePath)

	for _, sp := range p.Services {
		svc := svcs.Lookup(sp.Name)

		nextv := 1
		if svc != nil {
			if dv := svc.NextDeployVersion(); dv > 0 {
				nextv = dv
			}
		}

		if nextv != sp.Version {
			return nil, fmt.Errorf("service %s version changed to v%d since deploy request of v%d", sp.Name, nextv, sp.Version)
		}

		if err := destination.WriteDeployBundle(templateRoot, manifestRoot, sp); err != nil {
			return nil, fmt.Errorf("failed to write bundle: %w", err)
		}

		if err := destination.UpsertServiceStateForDeploy(manifestRoot, p.ProjectName, svc, sp); err != nil {
			return nil, fmt.Errorf("failed to update service state: %w", err)
		}
	}

	msg := fmt.Sprintf("Deploying %s", p.ProjectName)
	if len(p.Services) == 1 {
		msg = fmt.Sprintf("Deploying %s/%s", p.ProjectName, p.Services[0].Name)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := wt.AddGlob(s.bundlePath + "/*"); err != nil {
		return nil, fmt.Errorf("failed to add files to worktree: %w", err)
	}

	hash, err := wt.Commit(msg, &git.CommitOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	if err := repo.Push(&git.PushOptions{RemoteName: "origin", Auth: s.auth}); err != nil {
		return nil, fmt.Errorf("failed to push: %w", err)
	}

	return &destination.DeployResp{Hash: hash.String()}, nil
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
