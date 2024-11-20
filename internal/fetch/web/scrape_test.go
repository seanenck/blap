package web_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/retriever"
	"github.com/seanenck/blap/internal/fetch/web"
)

type mock struct {
	payload []byte
}

func (m *mock) Do(*http.Request) (*http.Response, error) {
	if m.payload != nil {
		return &http.Response{
			Body: io.NopCloser(bytes.NewReader(m.payload)),
		}, nil
	}
	return nil, nil
}

func (m *mock) Output(string, ...string) ([]byte, error) {
	return nil, nil
}

func TestScrape(t *testing.T) {
	client := &mock{}
	client.payload = []byte("1.2.3")
	r := &retriever.ResourceFetcher{}
	r.Backend = client
	if _, err := web.Scrape(r, fetch.Context{Name: "afa"}, core.WebMode{URL: "xyz", Scrape: &core.Filtered{Sort: "", Download: "ajfaeaijo", Filters: []string{"(.*?)"}}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("v0.1.2\n2.3.0")
	r.Backend = client
	if _, err := web.Scrape(r, fetch.Context{Name: "aaa"}, core.WebMode{URL: "xyz", Scrape: &core.Filtered{Sort: "semver", Download: "oijaoeja", Filters: []string{"(TEST?)"}}}); err == nil || err.Error() != "no tags found" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	o, err := web.Scrape(r, fetch.Context{Name: "aljfao"}, core.WebMode{URL: "a/xyz", Scrape: &core.Filtered{Sort: "semver", Download: "oijoefa/x", Filters: []string{"abc-([0-9.]*?).txt"}}})
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
	client = &mock{}
	client.payload = []byte("abc-v0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3.txt\nabc-v1.1.2.txt")
	r.Backend = client
	o, err = web.Scrape(r, fetch.Context{Name: "aaa"}, core.WebMode{URL: "a/xyz", Scrape: &core.Filtered{Sort: "semver", Download: "xx/s/{{ $.Vars.Tag }}/abc/{{ $.Name }}", Filters: []string{`{{ if ne $.Arch "invalidarch" }}abc-v([0-9.]*?).txt{{end}}`}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "v1.1.2" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "xx/s/v1.1.2/abc/aaa" || o.File != "aaa" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}
