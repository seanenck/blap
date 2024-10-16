package fetch_test

import (
	"testing"

	"github.com/seanenck/blap/internal/fetch"
)

func TestBranchValidate(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	if _, err := r.GitHubBranch(fetch.GitHubMode{}); err == nil || err.Error() != "branch is not properly set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := r.GitHubBranch(fetch.GitHubMode{Branch: &fetch.GitHubBranchMode{}}); err == nil || err.Error() != "branch required for branch mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := r.GitHubBranch(fetch.GitHubMode{Branch: &fetch.GitHubBranchMode{"aa"}}); err == nil || err.Error() != "project required for branch mode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestBranch(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	client := &mockClient{}
	client.Payload("{}")
	r.Requestor = client
	if _, err := r.GitHubBranch(fetch.GitHubMode{Branch: &fetch.GitHubBranchMode{"abc"}, Project: "xyz"}); err == nil || err.Error() != "invalid sha detected: " {
		t.Errorf("invalid error: %v", err)
	}
	client.Payload(`{"sha": "123456"}`)
	r.Requestor = client
	if _, err := r.GitHubBranch(fetch.GitHubMode{Branch: &fetch.GitHubBranchMode{"abc"}, Project: "xyz"}); err == nil || err.Error() != "invalid sha detected: 123456" {
		t.Errorf("invalid error: %v", err)
	}
	client.Payload(`{"sha": "1234567"}`)
	r.Requestor = client
	if _, err := r.GitHubBranch(fetch.GitHubMode{Branch: &fetch.GitHubBranchMode{"abc"}, Project: "xyz"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client.Payload(`{"sha": "12345678"}`)
	r.Requestor = client
	o, err := r.GitHubBranch(fetch.GitHubMode{Branch: &fetch.GitHubBranchMode{"abc"}, Project: "xyz"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset, nil")
	}
	if o.Tag != "1234567" || o.File != "1234567-abc.tar.gz" || o.URL != "https://github.com/xyz/archive/abc.tar.gz" {
		t.Errorf("invalid asset, %v", o)
	}
}
