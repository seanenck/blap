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

func TestScrapeValidate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	if _, err := web.Scrape(r, fetch.Context{Name: "abc"}, core.WebMode{}); err == nil || err.Error() != "scrape definition is nil" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := web.Scrape(r, fetch.Context{Name: "xyz"}, core.WebMode{Scrape: &core.Filtered{}}); err == nil || err.Error() != "no URL configured" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := web.Scrape(r, fetch.Context{Name: "xxx"}, core.WebMode{URL: "xyz", Scrape: &core.Filtered{}}); err == nil || err.Error() != "application scraping requires filters" {
		t.Errorf("invalid error: %v", err)
	}
	client := &mock{}
	client.payload = []byte("")
	r.Backend = client
	if _, err := web.Scrape(r, fetch.Context{Name: "j1o2i"}, core.WebMode{URL: "xyz", Scrape: &core.Filtered{Filters: []string{"TEST"}}}); err == nil || err.Error() != "no tags scraped" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestScrape(t *testing.T) {
	client := &mock{}
	client.payload = []byte("TESD\ta")
	r := &retriever.ResourceFetcher{}
	r.Backend = client
	if _, err := web.Scrape(r, fetch.Context{Name: "afa"}, core.WebMode{URL: "xyz", Scrape: &core.Filtered{Filters: []string{"(TEST?)"}}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("v0.1.2\n2.3.0")
	r.Backend = client
	if _, err := web.Scrape(r, fetch.Context{Name: "aaa"}, core.WebMode{URL: "xyz", Scrape: &core.Filtered{Filters: []string{"(TEST?)"}}}); err == nil || err.Error() != "no tags scraped" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	o, err := web.Scrape(r, fetch.Context{Name: "aljfao"}, core.WebMode{URL: "a/xyz", Scrape: &core.Filtered{Filters: []string{"abc-([0-9.]*?).txt"}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "v2.3.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "a/xyz" || o.File != "xyz" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
	client = &mock{}
	client.payload = []byte("abc-v0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3.txt\nabc-v1.1.2.txt")
	r.Backend = client
	o, err = web.Scrape(r, fetch.Context{Name: "aaa"}, core.WebMode{URL: "a/xyz", Scrape: &core.Filtered{Download: "xx/s/{{ $.Vars.Tag }}/abc/{{ $.Name }}", Filters: []string{`{{ if ne $.Arch "invalidarch" }}abc-v([0-9.]*?).txt{{end}}`}}})
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
	if _, err = web.Scrape(r, fetch.Context{}, core.WebMode{URL: "a/xyz", Scrape: &core.Filtered{Filters: []string{"TEST"}}}); err == nil || err.Error() != "context missing name" {
		t.Errorf("invalid error: %v", err)
	}
}
