package fetch_test

import (
	"testing"

	"github.com/seanenck/blap/internal/fetch"
)

func TestGitHubErrors(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	if _, err := r.GitHubRelease(fetch.GitHubMode{}); err == nil || err.Error() != "github mode requires a project" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz"}); err == nil || err.Error() != "release is not properly set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{}}); err == nil || err.Error() != "github mode requires an asset filter (regex)" {
		t.Errorf("invalid error: %v", err)
	}
	client := &mockClient{}
	client.payload = []byte("{}")
	r.Requestor = client
	if _, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{"zzz"}}); err == nil || err.Error() != "no assets found" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mockClient{}
	client.payload = []byte(`{"assets": [{"browser_download_url": "111"}, {"browser_download_url": "222"}]}`)
	r.Requestor = client
	if _, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{"zzz"}}); err == nil || err.Error() != "assets found but no tag" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mockClient{}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "111"}, {"browser_download_url": "222"}]}`)
	r.Requestor = client
	if _, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{"zzz"}}); err == nil || err.Error() != "unable to find asset, choices: [111 222]" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mockClient{}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "111"}, {"browser_download_url": "1222"}]}`)
	r.Requestor = client
	if _, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{"1"}}); err == nil || err.Error() != "multiple assets matched: 1222 (had: 111)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestGitHub(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	client := &mockClient{}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
	r.Requestor = client
	o, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{"1"}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset, nil")
	}
	if o.Tag != "123" || o.URL != "x/111" || o.File != "111" {
		t.Errorf("invalid asset: %s %s %s", o.Tag, o.URL, o.File)
	}
}

func TestGitHubTarball(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	client := &mockClient{}
	client.payload = []byte(`{"tag_name": "123a", "tarball_url": "afea/ddd", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
	r.Requestor = client
	o, err := r.GitHubRelease(fetch.GitHubMode{Project: "xyz", Release: &fetch.GitHubReleaseMode{"tarball"}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset, nil")
	}
	if o.Tag != "123a" || o.URL != "afea/ddd" || o.File != "ddd.tar.gz" {
		t.Errorf("invalid asset: %s %s %s", o.Tag, o.URL, o.File)
	}
}
