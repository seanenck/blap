// Package git gets git-tagged resources
package git

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
)

// Tagged gets a tagged (git tag) release
func Tagged(caller fetch.Retriever, ctx fetch.Context, a core.GitMode) (*core.Resource, error) {
	if a.Tagged == nil {
		return nil, errors.New("tagged definition is nil")
	}
	up := strings.TrimSpace(a.Repository)
	if up == "" {
		return nil, errors.New("no upstream for tagged mode")
	}
	dl := strings.TrimSpace(a.Tagged.Download)
	if dl == "" {
		dl = up
	}
	if len(a.Tagged.Filters) == 0 {
		return nil, errors.New("application lacks filters")
	}
	var re []*regexp.Regexp
	for _, r := range a.Tagged.Filters {
		r, err := ctx.CompileRegexp(r, nil)
		if err != nil {
			return nil, err
		}
		re = append(re, r)
	}
	out, err := caller.ExecuteCommand("git", "-c", "versionsort.suffix=-", "ls-remote", "--tags", "--sort=-v:refname", up)
	if err != nil {
		return nil, err
	}
	var tag string
	for _, line := range strings.Split(string(out), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		passed := true
		for _, r := range re {
			if r.MatchString(line) {
				passed = false
				break
			}
		}
		if passed {
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				return nil, fmt.Errorf("matching version line can not be parsed: %s", line)
			}
			tag = strings.TrimPrefix(parts[1], "refs/tags/")
			break
		}
	}
	if tag == "" {
		return nil, errors.New("no tags matched")
	}
	caller.Debug("found tag: %s\n", tag)
	tl, err := ctx.Templating(dl, &fetch.Template{Tag: tag})
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(tl)
	return &core.Resource{URL: url, File: filepath.Base(url), Tag: tag}, nil
}
