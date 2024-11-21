// Package web can scrape a page for versions and try to match filters
package web

import (
	"io"
	"regexp"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/filtered"
)

type scrapeFilterable struct {
	filtered.Base
}

func (s scrapeFilterable) Get(r fetch.Retriever, url string) ([]byte, error) {
	resp, err := r.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s scrapeFilterable) Match(r []*regexp.Regexp, line string) ([]string, error) {
	var results []string
	for _, re := range r {
		m := re.FindStringSubmatch(line)
		if len(m) == 0 {
			continue
		}
		matched := m[0]
		if len(m) > 1 {
			matched = m[1]
		}
		results = append(results, matched)
	}
	return results, nil
}

// Scrape will scrape a GET requested resource
func Scrape(caller fetch.Retriever, ctx fetch.Context, a core.WebMode) (*core.Resource, error) {
	b, err := filtered.NewBase(a.URL, a.Scrape, scrapeFilterable{})
	if err != nil {
		return nil, err
	}
	return b.Get(caller, ctx)
}
