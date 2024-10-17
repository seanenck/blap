package retriever_test

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/retriever"
)

type (
	mockClient struct {
		err     error
		payload []byte
		invalid bool
		req     *http.Request
	}
)

func (m *mockClient) Output(string, ...string) ([]byte, error) {
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

func TestGitHubFetchNotOK(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockClient{}
	client.invalid = true
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
	r.Backend = client
	if err := r.GitHubFetch("aaa", "bcy", struct{}{}); err == nil || !strings.Contains(err.Error(), "code: 404") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestDownload(t *testing.T) {
	r := &retriever.ResourceFetcher{}
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
	r.Backend = client
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
		r := &retriever.ResourceFetcher{}
		if set {
			r.SetToken("abc")
		}
		client := &mockClient{}
		client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
		r.Backend = client
		r.GitHubFetch("abc", "aaa", struct{}{})
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

func testIter(objs ...any) iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, o := range objs {
			if !yield(o) {
				return
			}
		}
	}
}

func TestProcess(t *testing.T) {
	f := retriever.ResourceFetcher{}
	if _, err := f.Process(fetch.Context{Name: ""}, testIter(nil)); err == nil || err.Error() != "name is required" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(fetch.Context{Name: "abc"}, testIter(nil, nil)); err == nil || err.Error() != "unknown mode for fetch processing" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(fetch.Context{Name: "abc"}, testIter(&types.GitHubMode{}, &types.GitMode{})); err == nil || err.Error() != "multiple modes enabled, only one allowed" {
		t.Errorf("invalid error: %v", err)
	}
	m := &mockClient{}
	f.Backend = m
	ctx := fetch.Context{Name: "a"}
	if _, err := f.Process(ctx, testIter(&types.GitHubMode{}, nil)); err == nil || err.Error() != "github mode set but not configured" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(&types.GitHubMode{Release: &types.GitHubReleaseMode{}, Branch: &types.GitHubBranchMode{}}, nil)); err == nil || err.Error() != "only one github mode is allowed" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(&types.GitHubMode{Branch: &types.GitHubBranchMode{}})); err == nil || err.Error() != "branch required for branch mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(&types.GitHubMode{Release: &types.GitHubReleaseMode{}}, nil)); err == nil || err.Error() != "release mode requires a project" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(nil, &types.GitMode{})); err == nil || err.Error() != "unknown git mode for fetch processing" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(nil, nil, &types.GitMode{Tagged: &types.GitTaggedMode{}}, nil)); err == nil || err.Error() != "no upstream for tagged mode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestGitHubFetch(t *testing.T) {
	m := &mockClient{}
	m.payload = []byte(`{"Num": 1}`)
	f := retriever.ResourceFetcher{}
	f.Backend = m
	if err := f.GitHubFetch("", "", nil); err == nil || err.Error() != "owner/repo and call must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := f.GitHubFetch("b", "", nil); err == nil || err.Error() != "owner/repo and call must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := f.GitHubFetch("", "a", nil); err == nil || err.Error() != "owner/repo and call must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := f.GitHubFetch("x", "a", nil); err == nil || err.Error() != "result object must be set" {
		t.Errorf("invalid error: %v", err)
	}
	type testType struct {
		Num int
	}
	obj := testType{}
	if err := f.GitHubFetch("a", "y", &obj); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if obj.Num != 1 {
		t.Errorf("invalid result: %d", obj.Num)
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	r := retriever.ResourceFetcher{}
	r.Context.Verbosity = 0
	r.Context.Writer = &buf
	r.Debug("abc")
	s := buf.String()
	if s != "" {
		t.Errorf("invalid buffer: %s", s)
	}
	r.Context.Verbosity = 100
	r.Debug("abc")
	s = buf.String()
	if s != "abc" {
		t.Errorf("invalid buffer: %s", s)
	}
}

func TestExecuteCommand(t *testing.T) {
	m := &mockClient{}
	m.payload = []byte("aaa")
	r := retriever.ResourceFetcher{}
	r.Backend = m
	o, err := r.ExecuteCommand("abc", "xyz")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if string(o) != "aaa" {
		t.Errorf("invalid result: %s", string(o))
	}
}
