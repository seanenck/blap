package filtered_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/filtered"
	"github.com/seanenck/blap/internal/fetch/retriever"
)

type (
	mockFilterable struct {
		payload []byte
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

func (s *mockFilterable) NewLine(line string) (string, error) {
	return line, nil
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
	if _, err := filtered.NewBase(filtered.RawString(""), nil, nil); err == nil || err.Error() != "filterable interface is nil" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := filtered.NewBase(filtered.RawString(""), nil, &mockFilterable{}); err == nil || err.Error() != "filter definition is nil" {
		t.Errorf("invalid error: %v", err)
	}
	data := &core.Filtered{}
	if _, err := filtered.NewBase(filtered.RawString(""), data, &mockFilterable{}); err == nil || err.Error() != "no upstream configured" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := filtered.NewBase(filtered.RawString("aaa"), data, &mockFilterable{}); err == nil || err.Error() != "no download URL configured" {
		t.Errorf("invalid error: %v", err)
	}
	data.Download = "aa"
	if _, err := filtered.NewBase(filtered.RawString("yao"), data, &mockFilterable{}); err == nil || err.Error() != "filters required" {
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
	b, err := filtered.NewBase(filtered.RawString("aaa"), data, &mockFilterable{})
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
	client.payload = []byte("abc-v2.3.0.txt\nabc-v4.3.0.txt\n\nabc-v3.2.1.txt\nabc-v1.2.3.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	mock.payload = client.payload
	b, err := filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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
	client.payload = []byte("abc-2.3.0.txt\nabc-2.31.0.txt\n\nabc-v2.10.0.txt\nabc-2.32.0.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	data.Sort = "semver"
	mock.payload = client.payload
	b, err := filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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
	client.payload = []byte("abc-2.3.0.txt\nabc-2.31.0.txt\n\nabc-v2.10.0txt\nabc-2.32.0.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	data.Sort = "rsemver"
	mock.payload = client.payload
	b, err := filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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
	client.payload = []byte("abc-2.3.0.txt\nabc-2.31.0.txt\n\nabc-2.10.0.txt\nabc-2.32.0.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	data.Sort = "rsort"
	mock.payload = client.payload
	b, err := filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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
	client.payload = []byte("abc-2.3.0.txt\nabc-2.31.0.txt\n\nabc-2.10.0.txt\nabc-2.32.0.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "oijoefa/x"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	data.Sort = "sort"
	mock.payload = client.payload
	b, err := filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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
	client.payload = []byte("abc-2.3.0.txt\nabc-2.31.0.txt\n\nabc-2.10.0.txt\nabc-2.32.0.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "{{ $.Vars.Source }}{{ $.Vars.Arguments }}/{{ $.Vars.Tag }}"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	data.Sort = "sort"
	mock.payload = client.payload
	b, err := filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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
	b, err = filtered.NewBase(filtered.RawString("a/xyz"), data, mock)
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

func TestFilterTemplate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	client := &mockFilterable{}
	client.payload = []byte("abc-2.3.0.txt\nabc-2.31.0.txt\n\nabc-2.10.0.txt\nabc-2.32.0.txt")
	r.Backend = client
	mock := &mockFilterable{}
	data := &core.Filtered{}
	data.Download = "{{ $.Vars.Source }}{{ $.Vars.Arguments }}/{{ $.Vars.Tag }}"
	data.Filters = append(data.Filters, "abc-(v?[0-9.]+).txt")
	data.Sort = "sort"
	mock.payload = client.payload
	defer os.Clearenv()
	t.Setenv("TEST", "aaa")
	b, err := filtered.NewBase(core.WebURL("{{ $.Getenv \"TEST\" }}/a/xyz"), data, mock)
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
		if o.URL != "aaa/a/xyz[]/2.32.0" || o.File != "2.32.0" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}

func TestRawString(t *testing.T) {
	s := filtered.RawString("A")
	if s.String() != "A" || s.CanTemplate() {
		t.Errorf("invalid raw string")
	}
}
