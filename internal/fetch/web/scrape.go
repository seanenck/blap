// Package web can scrape a page for versions and try to match filters
package web

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"golang.org/x/mod/semver"
)

// Scrape will scrape a GET requested resource
func Scrape(caller fetch.Retriever, ctx fetch.Context, a core.WebMode) (*core.Resource, error) {
	if a.Scrape == nil {
		return nil, errors.New("scrape definition is nil")
	}
	up := strings.TrimSpace(a.URL)
	if up == "" {
		return nil, errors.New("no URL configured")
	}
	dl := strings.TrimSpace(a.Scrape.Download)
	if dl == "" {
		dl = up
	}
	if len(a.Scrape.Filters) == 0 {
		return nil, errors.New("application scraping requires filters")
	}
	var re []*regexp.Regexp
	for _, r := range a.Scrape.Filters {
		r, err := ctx.CompileRegexp(r, nil)
		if err != nil {
			return nil, err
		}
		re = append(re, r)
	}
	resp, err := caller.Get(up)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var options []string
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		for _, r := range re {
			m := r.FindStringSubmatch(t)
			if len(m) == 0 {
				continue
			}
			matched := m[0]
			if len(m) > 1 {
				matched = m[1]
			}
			if !strings.HasPrefix(matched, "v") {
				matched = fmt.Sprintf("v%s", matched)
			}
			options = append(options, matched)
		}
	}
	if len(options) == 0 {
		return nil, errors.New("no tags scraped")
	}
	semver.Sort(options)
	tag := options[len(options)-1]
	caller.Debug("found tag: %s\n", tag)
	tl, err := ctx.Templating(dl, &fetch.Template{Tag: fetch.Version(tag)})
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(tl)
	return &core.Resource{URL: url, File: filepath.Base(url), Tag: tag}, nil
}
