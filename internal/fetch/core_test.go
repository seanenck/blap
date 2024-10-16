package fetch_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/fetch"
)

type (
	mockClient struct {
		err     error
		payload []byte
		invalid bool
		req     *http.Request
	}
	mockFetcher struct {
		mode string
	}
)

func (f *mockFetcher) GitHub(fetch.GitHubMode) (*asset.Resource, error) {
	f.mode = "gh"
	return nil, nil
}

func (f *mockFetcher) Tagged(fetch.TaggedMode) (*asset.Resource, error) {
	f.mode = "tag"
	return nil, nil
}

func (f *mockFetcher) Branch(fetch.BranchMode) (*asset.Resource, error) {
	f.mode = "branch"
	return nil, nil
}

func (m *mockClient) Output() ([]byte, error) {
	return m.payload, m.err
}

func (m *mockClient) Do(r *http.Request) (*http.Response, error) {
	m.req = r
	length := len(m.payload)
	if length > 0 {
		resp := &http.Response{}
		resp.Body = io.NopCloser(bytes.NewBuffer(m.payload))
		resp.ContentLength = int64(length)
		resp.StatusCode = http.StatusOK
		if m.invalid {
			resp.StatusCode = http.StatusNotFound
		}
		return resp, nil
	}
	return nil, m.err
}

func (m *mockClient) Payload(data string) {
	m.payload = []byte(data)
}

func TestGitHubWrapperError(t *testing.T) {
	w := fetch.GitHubWrapperError{}
	if w.Error() != "code: 0" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.URL = "a"
	if w.Error() != "code: 0\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Status = "x"
	if w.Error() != "code: 0\nstatus: x\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Body = []byte("[")
	if w.Error() != "code: 0\nstatus: x\nunmarshal: unexpected end of JSON input\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Body = []byte("{}")
	if w.Error() != "code: 0\nstatus: x\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Body = []byte(`{"message": "mess", "documentation_url": "xxx"}`)
	if w.Error() != "code: 0\ndoc: xxx\nmessage: mess\nstatus: x\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
}

func TestNotOKStatus(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	client := &mockClient{}
	client.invalid = true
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
	r.Requestor = client
	if _, err := r.GitHub(fetch.GitHubMode{Project: "xyz", Asset: "1"}); err == nil || !strings.Contains(err.Error(), "code: 404") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestDownload(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	if _, err := r.Download(true, "", ""); err == nil || err.Error() != "source (url) and destination (path) required" {
		t.Errorf("invalid error error: %v", err)
	}
	if _, err := r.Download(true, "a", ""); err == nil || err.Error() != "source (url) and destination (path) required" {
		t.Errorf("invalid error error: %v", err)
	}
	if _, err := r.Download(true, "", "d"); err == nil || err.Error() != "source (url) and destination (path) required" {
		t.Errorf("invalid error error: %v", err)
	}
	client := &mockClient{}
	client.payload = []byte(`abc`)
	r.Requestor = client
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	path := filepath.Join("testdata", "file")
	did, err := r.Download(true, "d", path)
	if !did || err != nil {
		t.Errorf("did wrong or error: %v", err)
	}
	os.WriteFile(path, []byte{}, 0o644)
	did, err = r.Download(true, "d", path)
	if did || err != nil {
		t.Errorf("did wrong or error: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "" {
		t.Error("should be empty")
	}
	os.Remove(path)
	did, err = r.Download(false, "d", path)
	if !did || err != nil {
		t.Errorf("did wrong or error: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "abc" {
		t.Error("should be empty")
	}
}

func TestTokenHandling(t *testing.T) {
	os.Clearenv()
	checkToken := func(unset, set bool, expect string) {
		defer os.Clearenv()
		r := &fetch.ResourceFetcher{}
		if set {
			r.SetToken("abc")
		}
		client := &mockClient{}
		client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
		r.Requestor = client
		r.GitHub(fetch.GitHubMode{Project: "xyz", Asset: "1"})
		h, ok := client.req.Header["Authorization"]
		if unset {
			if ok {
				t.Error("token should be unset")
			}
		} else {
			if expect != "[]" {
				if !ok {
					t.Error("token should be set")
				}
			}
		}
		if fmt.Sprintf("%v", h) != expect {
			t.Errorf("%s != %s", h, expect)
		}
	}
	checkToken(true, false, "[]")
	checkToken(false, false, "[]")
	checkToken(false, true, "[token abc]")
	t.Setenv("GITHUB_TOKEN", "xyz")
	t.Setenv("BLAP_GITHUB_TOKEN", "xyz1")
	checkToken(false, false, "[token xyz1]")
	t.Setenv("BLAP_GITHUB_TOKEN", "xyz1")
	checkToken(false, true, "[token xyz1]")
	t.Setenv("GITHUB_TOKEN", "xyz")
	checkToken(false, false, "[token xyz]")
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	tokenFile := filepath.Join("testdata", "token")
	os.WriteFile(tokenFile, []byte("xxx"), 0o644)
	t.Setenv("GITHUB_TOKEN", tokenFile)
	checkToken(false, false, "[token xxx]")
}

func TestProcess(t *testing.T) {
	f := fetch.ResourceFetcher{}
	if _, err := f.Process(&mockFetcher{}, nil, nil, nil); err == nil || err.Error() != "unknown mode for fetch processing" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(&mockFetcher{}, &fetch.GitHubMode{}, &fetch.TaggedMode{}, nil); err == nil || err.Error() != "multiple modes enabled, only one allowed" {
		t.Errorf("invalid error: %v", err)
	}
	m := &mockFetcher{}
	if _, err := f.Process(m, &fetch.GitHubMode{}, nil, nil); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "gh" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	if _, err := f.Process(m, nil, nil, &fetch.BranchMode{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "branch" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	if _, err := f.Process(m, nil, &fetch.TaggedMode{}, nil); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "tag" {
		t.Errorf("invalid mode: %s", m.mode)
	}
}
