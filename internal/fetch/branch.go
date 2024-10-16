// Package fetch can handle git branch sources
package fetch

import (
	"errors"
	"fmt"

	"github.com/seanenck/blap/internal/asset"
)

type (
	// GitHubCommit is commit information from github for a repo
	GitHubCommit struct {
		Sha string `json:"sha"`
	}
	// BranchMode will enable a repository+branch to pull a tarball
	BranchMode struct {
		Project string
		Branch  string
	}
)

// Branch will get an asset from a branch
func (r *ResourceFetcher) Branch(a BranchMode) (*asset.Resource, error) {
	if a.Branch == "" {
		return nil, errors.New("branch required for branch mode")
	}
	if a.Project == "" {
		return nil, errors.New("project required for branch mode")
	}
	commit, err := fetchData[GitHubCommit](r, a.Project, fmt.Sprintf("commits/%s", a.Branch))
	if err != nil {
		return nil, err
	}
	if len(commit.Sha) < 7 {
		return nil, fmt.Errorf("invalid sha detected: %s", commit.Sha)
	}
	tag := commit.Sha[0:7]
	r.Context.LogDebug("found sha: %s\n", tag)
	return &asset.Resource{URL: fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", a.Project, a.Branch), File: fmt.Sprintf("%s-%s.tar.gz", tag, a.Branch), Tag: tag}, nil
}
