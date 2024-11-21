// Package filtered handles components that are 'raw' lines of text that need to become a version
package filtered

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"golang.org/x/mod/semver"
)

type (
	// Base are wrappers for filtering components
	Base struct {
		upstream string
		data     *core.Filtered
	}
)

// NewBase creates a new baseline for filtering
func NewBase(upstream string, d *core.Filtered) Base {
	return Base{upstream: upstream, data: d}
}

// Upstream returns the underlying upstream
func (b Base) Upstream() string {
	return b.upstream
}

// Definition returns the filter definition
func (b Base) Definition() *core.Filtered {
	return b.data
}

// Get handles common filtered commands that have return lists of semver versions
func Get(r fetch.Retriever, ctx fetch.Context, filterable fetch.Filterable) (*core.Resource, error) {
	f := filterable.Definition()
	if f == nil {
		return nil, errors.New("filter definition is nil")
	}
	up := strings.TrimSpace(filterable.Upstream())
	if up == "" {
		return nil, errors.New("no upstream configured")
	}
	dl := strings.TrimSpace(f.Download)
	if dl == "" {
		return nil, errors.New("no download URL configured")
	}
	if len(f.Filters) == 0 {
		return nil, errors.New("filters required")
	}
	const (
		reversePrefix = "r"
		rSemVerType   = reversePrefix + "semver"
		rSortType     = reversePrefix + "sort"
		sortType      = "sort"
		semVerType    = "semver"
	)
	isSemVer := false
	isSort := false
	switch f.Sort {
	case "":
		break
	case rSortType, sortType:
		isSort = true
	case rSemVerType, semVerType:
		isSemVer = true
	default:
		return nil, fmt.Errorf("unknown sort type: %s", f.Sort)
	}
	var re []*regexp.Regexp
	for _, r := range f.Filters {
		r, err := ctx.CompileRegexp(r, nil)
		if err != nil {
			return nil, err
		}
		re = append(re, r)
	}
	b, err := filterable.Get(r, up)
	if err != nil {
		return nil, err
	}
	var options []string
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		matches, err := filterable.Match(re, t)
		if err != nil {
			return nil, err
		}
		for _, opt := range matches {
			matched := opt
			if isSemVer {
				if !strings.HasPrefix(matched, "v") {
					matched = fmt.Sprintf("v%s", matched)
				}
				if !semver.IsValid(matched) {
					r.Debug("semver found an invalid match: %s\n", matched)
					continue
				}
			}
			options = append(options, matched)
		}
	}
	if len(options) == 0 {
		return nil, errors.New("no tags found")
	}
	if isSemVer {
		semver.Sort(options)
	} else if isSort {
		sort.Strings(options)
	}
	// this seems counter to what it should be but semver/sort should be defaults to get the newest version
	// reversing should be a backup
	if f.Sort != "" && !strings.HasPrefix(f.Sort, reversePrefix) {
		slices.Reverse(options)
	}
	tag := options[0]
	r.Debug("found tag: %s\n", tag)
	tl, err := ctx.Templating(dl, &fetch.Template{Tag: fetch.Version(tag)})
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(tl)
	return &core.Resource{URL: url, File: filepath.Base(url), Tag: tag}, nil
}
