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
		payload []byte
		matches []string
		args    []string
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

func (s *mockFilterable) Get(r fetch.Retriever, url string) ([]byte, error) {
	return s.payload, nil
}

func (s *mockFilterable) Match(r []*regexp.Regexp, line string) ([]string, error) {
	return s.matches, nil
}

func (s *mockFilterable) Arguments() []string {
	return s.args
}

func TestNewBaseInvalid(t *testing.T) {
	b := filtered.Base{}
	if _, err := b.Get(&retriever.ResourceFetcher{}, fetch.Context{}); err == nil || err.Error() != "invalid base is not configured" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewBaseValidate(t *testing.T) {
	if _, err := filtered.NewBase("", nil, nil); err == nil || err.Error() != "filterable interface is nil" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := filtered.NewBase("", nil, &mockFilterable{}); err == nil || err.Error() != "filter definition is nil" {
		t.Errorf("invalid error: %v", err)
	}
	data := &core.Filtered{}
	if _, err := filtered.NewBase("", data, &mockFilterable{}); err == nil || err.Error() != "no upstream configured" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := filtered.NewBase("aaa", data, &mockFilterable{}); err == nil || err.Error() != "no download URL configured" {
		t.Errorf("invalid error: %v", err)
	}
	data.Download = "aa"
	if _, err := filtered.NewBase("yao", data, &mockFilterable{}); err == nil || err.Error() != "filters required" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestGetErrors(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	data := &core.Filtered{}
	data.Download = "xxx"
	data.Filters = append(data.Filters, "aa")
	client := &mockFilterable{}
	client.payload = []byte("")
	r.Backend = client
	b, err := filtered.NewBase("aaa", data, &mockFilterable{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := b.Get(r, fetch.Context{Name: "j1o2i"}); err == nil || err.Error() != "no tags found" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := b.Get(r, fetch.Context{}); err == nil || err.Error() != "context missing name" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestBasicFilters(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-([0-9.]*?).txt")
	mock.matches = []string{"v2.3.0", "v4.3.0", "v3.2.1", "v1.2.3"}
	mock.payload = client.payload
	b, err := filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err := b.Get(r, fetch.Context{Name: "aljfao"})
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
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-([0-9.]*?).txt")
	data.Sort = "semver"
	mock.matches = []string{"2.3.0", "2.31.0", "v2.10.0", "2.32.0"}
	mock.payload = client.payload
	b, err := filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err := b.Get(r, fetch.Context{Name: "aljfao"})
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
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-([0-9.]*?).txt")
	data.Sort = "rsemver"
	mock.matches = []string{"2.3.0", "2.31.0", "v2.10.0", "2.32.0"}
	mock.payload = client.payload
	b, err := filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err := b.Get(r, fetch.Context{Name: "aljfao"})
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
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-([0-9.]*?).txt")
	data.Sort = "rsort"
	mock.matches = []string{"2.3.0", "2.31.0", "2.10.0", "2.32.0"}
	mock.payload = client.payload
	b, err := filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err := b.Get(r, fetch.Context{Name: "aljfao"})
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
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-([0-9.]*?).txt")
	data.Sort = "sort"
	mock.matches = []string{"2.3.0", "2.31.0", "2.10.0", "2.32.0"}
	mock.payload = client.payload
	b, err := filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err := b.Get(r, fetch.Context{Name: "aljfao"})
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

func TestTemplate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "{{ $.Vars.Source }}{{ $.Vars.Arguments }}/{{ $.Vars.Tag }}"
	data.Filters = append(data.Filters, "abc-([0-9.]*?).txt")
	data.Sort = "sort"
	mock.matches = []string{"2.3.0", "2.31.0", "2.10.0", "2.32.0"}
	mock.payload = client.payload
	b, err := filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err := b.Get(r, fetch.Context{Name: "aljfao"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.URL != "a/xyz[]/2.32.0" || o.File != "2.32.0" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
	mock.args = []string{"111", "222"}
	b, err = filtered.NewBase("a/xyz", data, mock)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err = b.Get(r, fetch.Context{Name: "aljfao"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.URL != "a/xyz[111 222]/2.32.0" || o.File != "2.32.0" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}

func TestMatchLine(t *testing.T) {
	matches := filtered.MatchLine(nil, "")
	if len(matches) != 0 {
		t.Error("should have no matches")
	}
	matches = filtered.MatchLine([]*regexp.Regexp{}, "")
	if len(matches) != 0 {
		t.Error("should have no matches")
	}
	matches = filtered.MatchLine([]*regexp.Regexp{
		regexp.MustCompile(".*"),
	}, "")
	if len(matches) != 1 {
		t.Error("should have matches")
	}
	matches = filtered.MatchLine([]*regexp.Regexp{
		regexp.MustCompile("[0-9.]*"),
		regexp.MustCompile("[0-9.]*"),
	}, "0.1\n0.2")
	if len(matches) != 2 {
		t.Error("should have matches")
	}
}