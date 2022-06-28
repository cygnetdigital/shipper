// Package source provides primitives for accessing service configuration,
// docker builds, and CI checks.
package source

import (
	"time"

	"github.com/cygnetdigital/shipper"
)

// Source encapsulates service configuration, docker builds and CI checks for
// a given commit.
type Source struct {
	// Name of the source project
	ProjectName string

	// The resolved git reference
	Ref *Ref

	// Project configuration for a fully resolved ref
	Project *shipper.Project

	// Services with builds queued/running/completed
	Services Services

	// Indicates that checks are still running
	ChecksRunning bool

	// Indicates that the services have all been built
	ChecksComplete bool
}

// GitHash represents a git hash.
type GitHash string

// Short version of the hash
func (h GitHash) Short() string {
	return string(h)[:7]
}

// Ref is a resolved reference to a PR, commit or branch.
// It includes information about how it was merged into main.
type Ref struct {
	// Type of reference that was resolved
	GivenRef string

	// Hash of the commit in the main branch this ref resolves to. This will be
	// the merge commit hash for a PR. If the PR has not been merged, this will
	// be empty.
	CommitHash GitHash

	// Details about the PullRequest
	PullRequest *PullRequest
}

// PullRequest details
type PullRequest struct {
	Title      string
	Number     int
	URL        string
	HeadCommit GithubCommit
	BaseCommit GithubCommit

	Merged           bool
	MergedAt         time.Time
	MergeCommitHash  GitHash
	MergedByUsername string
}

// GithubCommit ...
type GithubCommit struct {
	Hash     GitHash
	Ref      string
	Username string
}
