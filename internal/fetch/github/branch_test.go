package github_test

import (
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/github"
	"github.com/seanenck/blap/internal/fetch/retriever"
)

func TestBranchValidate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	r.Backend = &mock{}
	if _, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{}); err == nil || err.Error() != "branch is not properly set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{Branch: &core.GitHubBranchMode{}}); err == nil || err.Error() != "branch required for branch mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{Branch: &core.GitHubBranchMode{Name: "aa"}}); err == nil || err.Error() != "project required for branch mode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestBranch(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mock{}
	client.payload = []byte("{}")
	r.Backend = client
	if _, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{Branch: &core.GitHubBranchMode{Name: "abc"}, Project: "xyz"}); err == nil || err.Error() != "invalid sha detected: " {
		t.Errorf("invalid error: %v", err)
	}
	client.payload = []byte(`{"sha": "123456"}`)
	if _, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{Branch: &core.GitHubBranchMode{Name: "abc"}, Project: "xyz"}); err == nil || err.Error() != "invalid sha detected: 123456" {
		t.Errorf("invalid error: %v", err)
	}
	client.payload = []byte(`{"sha": "1234567"}`)
	if _, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{Branch: &core.GitHubBranchMode{Name: "abc"}, Project: "xyz"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client.payload = []byte(`{"sha": "12345678"}`)
	o, err := github.Branch(r, fetch.Context{Name: "br"}, core.GitHubMode{Branch: &core.GitHubBranchMode{Name: "abc"}, Project: "xyz"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset, nil")
	} else {
		if o.Tag != "1234567" || o.File != "1234567-abc.tar.gz" || o.URL != "https://github.com/xyz/archive/abc.tar.gz" {
			t.Errorf("invalid asset, %v", o)
		}
	}
}
