package retriever_test

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/core"
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
	mockFilterable struct {
		data     *core.Filtered
		upstream string
		payload  []byte
		matches  []string
	}
)

func (s *mockFilterable) Do(*http.Request) (*http.Response, error) {
	if s.payload != nil {
		return &http.Response{
			Body: io.NopCloser(bytes.NewReader(s.payload)),
		}, nil
	}
	return nil, nil
}

func (s *mockFilterable) Output(string, ...string) ([]byte, error) {
	return s.payload, nil
}

func (s *mockFilterable) Upstream() string {
	return s.upstream
}

func (s *mockFilterable) Get(r fetch.Retriever, url string) ([]byte, error) {
	return s.payload, nil
}

func (s *mockFilterable) Definition() *core.Filtered {
	return s.data
}

func (s *mockFilterable) Match(r []*regexp.Regexp, line string) ([]string, error) {
	return s.matches, nil
}

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

func TestSetConnections(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	r := &retriever.ResourceFetcher{}
	client := &mockClient{}
	client.payload = []byte(`{"tag_name": "123", "assets": [{"browser_download_url": "x/111"}, {"browser_download_url": "222"}]}`)
	r.Backend = client
	r.GitHubFetch("abc", "aaa", struct{}{})
	h, ok := client.req.Header["Authorization"]
	if ok || fmt.Sprintf("%v", h) != "[]" {
		t.Errorf("invalid header: %v", h)
	}
	if client.req.URL.String() != "https://api.github.com/repos/abc/aaa" {
		t.Errorf("invalid url: %v", client.req.URL)
	}
	os.Clearenv()
	c := core.Connections{}
	r.SetConnections(c)
	c.GitHub.Token = "xyz"
	r.SetConnections(c)
	r.Backend = client
	r.GitHubFetch("abc", "aaa", struct{}{})
	h, ok = client.req.Header["Authorization"]
	if !ok || fmt.Sprintf("%v", h) != "[token xyz]" {
		t.Errorf("invalid header: %v", h)
	}
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
	if _, err := f.Process(fetch.Context{Name: "abc"}, testIter(&core.GitHubMode{}, &core.GitMode{})); err == nil || err.Error() != "multiple modes enabled, only one allowed" {
		t.Errorf("invalid error: %v", err)
	}
	m := &mockClient{}
	f.Backend = m
	ctx := fetch.Context{Name: "a"}
	if _, err := f.Process(ctx, testIter(&core.GitHubMode{}, nil)); err == nil || err.Error() != "github mode set but not configured" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(&core.GitHubMode{Release: &core.GitHubReleaseMode{}, Branch: &core.GitHubBranchMode{}}, nil)); err == nil || err.Error() != "only one github mode is allowed" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(&core.GitHubMode{Branch: &core.GitHubBranchMode{}})); err == nil || err.Error() != "branch required for branch mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(&core.GitHubMode{Release: &core.GitHubReleaseMode{}}, nil)); err == nil || err.Error() != "release mode requires a project" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(nil, &core.GitMode{})); err == nil || err.Error() != "unknown git mode for fetch processing" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(nil, nil, &core.GitMode{Tagged: &core.Filtered{}}, nil)); err == nil || err.Error() != "no upstream configured" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := f.Process(ctx, testIter(nil, nil, &core.WebMode{Scrape: &core.Filtered{}}, nil)); err == nil || err.Error() != "no upstream configured" {
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

func TestFilteredValidate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	mock := &mockFilterable{}
	if _, err := r.Filtered(fetch.Context{Name: "abc"}, mock); err == nil || err.Error() != "filter definition is nil" {
		t.Errorf("invalid error: %v", err)
	}
	mock.data = &core.Filtered{}
	if _, err := r.Filtered(fetch.Context{Name: "xyz"}, mock); err == nil || err.Error() != "no upstream configured" {
		t.Errorf("invalid error: %v", err)
	}
	mock.upstream = "aaa"
	if _, err := r.Filtered(fetch.Context{Name: "xyz"}, mock); err == nil || err.Error() != "no download URL configured" {
		t.Errorf("invalid error: %v", err)
	}
	mock.data.Download = "xxx"
	if _, err := r.Filtered(fetch.Context{Name: "xxx"}, mock); err == nil || err.Error() != "filters required" {
		t.Errorf("invalid error: %v", err)
	}
	mock.data.Filters = append(mock.data.Filters, "aa")
	client := &mockFilterable{}
	client.payload = []byte("")
	r.Backend = client
	if _, err := r.Filtered(fetch.Context{Name: "j1o2i"}, mock); err == nil || err.Error() != "no tags found" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := r.Filtered(fetch.Context{}, mock); err == nil || err.Error() != "context missing name" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestFiltered(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.matches = []string{"v2.3.0", "v2.3.0"}
	mock.payload = client.payload
	o, err := r.Filtered(fetch.Context{Name: "aljfao"}, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "v2.3.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "oijoefa/x" || o.File != "x" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}

func TestFilteredSemVer(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.data.SemVer = true
	mock.matches = []string{"v2.3.0", "v1.1.1"}
	mock.payload = client.payload
	o, err := r.Filtered(fetch.Context{Name: "aljfao"}, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "v2.3.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "oijoefa/x" || o.File != "x" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}
