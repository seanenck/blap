package filtered_test

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/filtered"
	"github.com/seanenck/blap/internal/fetch/retriever"
)

type (
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

func TestNewBase(t *testing.T) {
	b := filtered.Base{}
	if b.Upstream() != "" || b.Definition() != nil {
		t.Errorf("invalid base: %v", b)
	}
	b = filtered.NewBase("abc", nil)
	if b.Upstream() != "abc" || b.Definition() != nil {
		t.Errorf("invalid base: %v", b)
	}
	b = filtered.NewBase("xyz", &core.Filtered{})
	if b.Upstream() != "xyz" || b.Definition() == nil {
		t.Errorf("invalid base: %v", b)
	}
}

func TestValidate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	mock := &mockFilterable{}
	if _, err := filtered.Get(r, fetch.Context{Name: "abc"}, mock); err == nil || err.Error() != "filter definition is nil" {
		t.Errorf("invalid error: %v", err)
	}
	mock.data = &core.Filtered{}
	if _, err := filtered.Get(r, fetch.Context{Name: "xyz"}, mock); err == nil || err.Error() != "no upstream configured" {
		t.Errorf("invalid error: %v", err)
	}
	mock.upstream = "aaa"
	if _, err := filtered.Get(r, fetch.Context{Name: "xyz"}, mock); err == nil || err.Error() != "no download URL configured" {
		t.Errorf("invalid error: %v", err)
	}
	mock.data.Download = "xxx"
	if _, err := filtered.Get(r, fetch.Context{Name: "xxx"}, mock); err == nil || err.Error() != "filters required" {
		t.Errorf("invalid error: %v", err)
	}
	mock.data.Filters = append(mock.data.Filters, "aa")
	client := &mockFilterable{}
	client.payload = []byte("")
	r.Backend = client
	if _, err := filtered.Get(r, fetch.Context{Name: "j1o2i"}, mock); err == nil || err.Error() != "no tags found" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := filtered.Get(r, fetch.Context{}, mock); err == nil || err.Error() != "context missing name" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestBasicFilters(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.matches = []string{"v2.3.0", "v4.3.0", "v3.2.1", "v1.2.3"}
	mock.payload = client.payload
	o, err := filtered.Get(r, fetch.Context{Name: "aljfao"}, mock)
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

func TestSemVer(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.data.Sort = "semver"
	mock.matches = []string{"2.3.0", "2.31.0", "v2.10.0", "2.32.0"}
	mock.payload = client.payload
	o, err := filtered.Get(r, fetch.Context{Name: "aljfao"}, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "v2.32.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "oijoefa/x" || o.File != "x" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}

func TestSemVerReverse(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.data.Sort = "rsemver"
	mock.matches = []string{"2.3.0", "2.31.0", "v2.10.0", "2.32.0"}
	mock.payload = client.payload
	o, err := filtered.Get(r, fetch.Context{Name: "aljfao"}, mock)
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

func TestSortReverse(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.data.Sort = "rsort"
	mock.matches = []string{"2.3.0", "2.31.0", "2.10.0", "2.32.0"}
	mock.payload = client.payload
	o, err := filtered.Get(r, fetch.Context{Name: "aljfao"}, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "2.10.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "oijoefa/x" || o.File != "x" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}

func TestSort(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	mock.upstream = "a/xyz"
	mock.data = &core.Filtered{}
	mock.data.Download = "oijoefa/x"
	mock.data.Filters = append(mock.data.Filters, "abc-([0-9.]*?).txt")
	mock.data.Sort = "sort"
	mock.matches = []string{"2.3.0", "2.31.0", "2.10.0", "2.32.0"}
	mock.payload = client.payload
	o, err := filtered.Get(r, fetch.Context{Name: "aljfao"}, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "2.32.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "oijoefa/x" || o.File != "x" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}
