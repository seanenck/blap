package github_test

import (
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/github"
	"github.com/seanenck/blap/internal/fetch/retriever"
)

func TestGitHubErrors(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{}); err == nil || err.Error() != "release mode requires a project" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz"}); err == nil || err.Error() != "release is not properly set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{}}); err == nil || err.Error() != "release mode requires an asset filter (regex)" {
		t.Errorf("invalid error: %v", err)
	}
	client := &mock{}
	client.payload = []byte("{}")
	r.Backend = client
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{Asset: "zzz"}}); err == nil || err.Error() != "no assets found" {
		t.Errorf("invalid error: %v", err)
	}
	client.payload = []byte(`{"assets": [{"browser_download_url": "111"}, {"browser_download_url": "222"}]}`)
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{Asset: "zzz"}}); err == nil || err.Error() != "assets found but no tag" {
		t.Errorf("invalid error: %v", err)
	}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "111"}, {"browser_download_url": "222"}]}`)
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{Asset: "zzz"}}); err == nil || err.Error() != "unable to find asset, choices:\n  -> 111\n  -> 222" {
		t.Errorf("invalid error: %v", err)
	}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "111"}, {"browser_download_url": "1222"}]}`)
	if _, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{Asset: "1"}}); err == nil || err.Error() != "multiple assets matched: 1222 (had: 111)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestGitHub(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mock{}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/abc/123.test.tar"}, {"browser_download_url": "222"}]}`)
	r.Backend = client
	o, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{Asset: "{{ $.Vars.Tag }}"}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset, nil")
	} else {
		if o.Tag != "123" || o.URL != "x/abc/123.test.tar" || o.File != "123.test.tar" {
			t.Errorf("invalid asset: %s %s %s", o.Tag, o.URL, o.File)
		}
	}
}

func TestGitHubTarball(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mock{}
	client.payload = []byte(`{"tag_name": "123a", "tarball_url": "afea/ddd", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
	r.Backend = client
	o, err := github.Release(r, fetch.Context{Name: "xyz"}, core.GitHubMode{Project: "xyz", Release: &core.GitHubReleaseMode{Asset: "tarball"}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset, nil")
	} else {
		if o.Tag != "123a" || o.URL != "afea/ddd" || o.File != "ddd.tar.gz" {
			t.Errorf("invalid asset: %s %s %s", o.Tag, o.URL, o.File)
		}
	}
}
