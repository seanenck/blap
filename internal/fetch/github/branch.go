// Package github can handle git branch sources
package github

import (
	"errors"
	"fmt"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/logging"
)

// Branch will get an asset from a branch
func Branch(caller fetch.Retriever, _ fetch.Context, a core.GitHubMode) (*core.Resource, error) {
	if a.Branch == nil {
		return nil, errors.New("branch is not properly set")
	}
	if a.Branch.Name == "" {
		return nil, errors.New("branch required for branch mode")
	}
	if a.Project == "" {
		return nil, errors.New("project required for branch mode")
	}
	type Commit struct {
		Sha string `json:"sha"`
	}
	commit := Commit{}
	if err := caller.GitHubFetch(a.Project, fmt.Sprintf("commits/%s", a.Branch.Name), &commit); err != nil {
		return nil, err
	}
	if len(commit.Sha) < 7 {
		return nil, fmt.Errorf("invalid sha detected: %s", commit.Sha)
	}
	tag := commit.Sha[0:7]
	caller.Debug(logging.GitHubCategory, "found sha: %s\n", tag)
	return &core.Resource{URL: fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", a.Project, a.Branch.Name), File: fmt.Sprintf("%s-%s.tar.gz", tag, a.Branch.Name), Tag: tag}, nil
}
