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
	"github.com/seanenck/blap/internal/logging"
	"golang.org/x/mod/semver"
)

type (
	// RawString is a string that is not templated and just...a string
	RawString string
	// Stringable allows for filtering input strings via templating
	Stringable interface {
		String() string
		CanTemplate() bool
	}
	// Base are wrappers for filtering components
	Base struct {
		upstream struct {
			value    string
			template bool
		}
		data       *core.Filtered
		filterable fetch.Filterable
		valid      bool
		args       []string
	}
)

// String returns the underlying string
func (r RawString) String() string {
	return string(r)
}

// CanTemplate indicates if the type can template (this one cannot)
func (r RawString) CanTemplate() bool {
	return false
}

// NewBase creates a new baseline for filtering
func NewBase(upstream Stringable, d *core.Filtered, i fetch.Filterable) (Base, error) {
	if i == nil {
		return Base{}, errors.New("filterable interface is nil")
	}
	if d == nil {
		return Base{}, errors.New("filter definition is nil")
	}
	val := upstream.String()
	if val == "" {
		return Base{}, errors.New("no upstream configured")
	}
	dl := strings.TrimSpace(d.Download)
	if dl == "" {
		return Base{}, errors.New("no download URL configured")
	}
	if len(d.Filters) == 0 {
		return Base{}, errors.New("filters required")
	}
	base := Base{data: d, filterable: i, valid: true, args: i.Arguments()}
	base.upstream.value = val
	base.upstream.template = upstream.CanTemplate()
	return base, nil
}

// Get handles common filtered commands that have return lists of semver versions
func (b Base) Get(r fetch.Retriever, ctx fetch.Context) (*core.Resource, error) {
	if !b.valid {
		return nil, errors.New("invalid base is not configured")
	}
	filterable := b.filterable
	const (
		reversePrefix = "r"
		rSemVerType   = reversePrefix + "semver"
		rSortType     = reversePrefix + "sort"
		sortType      = "sort"
		semVerType    = "semver"
	)
	isSemVer := false
	isSort := false
	switch b.data.Sort {
	case "":
		break
	case rSortType, sortType:
		isSort = true
	case rSemVerType, semVerType:
		isSemVer = true
	default:
		return nil, fmt.Errorf("unknown sort type: %s", b.data.Sort)
	}
	var re []*regexp.Regexp
	for _, r := range b.data.Filters {
		r, err := ctx.CompileRegexp(r, nil)
		if err != nil {
			return nil, err
		}
		re = append(re, r)
	}
	up := b.upstream.value
	if b.upstream.template {
		updated, err := ctx.Templating(up, struct{}{})
		if err != nil {
			return nil, err
		}
		up = updated
	}

	r.Debug(logging.FilteringCategory, "url: %s\n", up)
	data, err := filterable.Get(r, up)
	if err != nil {
		return nil, err
	}
	var options []string
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		parsed, err := filterable.NewLine(t)
		if err != nil {
			return nil, err
		}
		t = parsed
		for _, opt := range matchLine(re, t) {
			matched := opt
			if isSemVer {
				if !strings.HasPrefix(matched, "v") {
					matched = fmt.Sprintf("v%s", matched)
				}
				if !semver.IsValid(matched) {
					r.Debug(logging.FilteringCategory, "semver found an invalid match: %s\n", matched)
					continue
				}
			}
			r.Debug(logging.FilteringCategory, "matched version: %s\n", matched)
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
	if b.data.Sort != "" && !strings.HasPrefix(b.data.Sort, reversePrefix) {
		slices.Reverse(options)
	}
	tag := options[0]
	r.Debug(logging.FilteringCategory, "found tag: %s\n", tag)
	type filterTemplate struct {
		*fetch.Template
		Source    string
		Arguments []string
	}
	t := filterTemplate{}
	t.Template = &fetch.Template{Tag: core.Version(tag)}
	t.Source = up
	t.Arguments = b.args
	tl, err := ctx.Templating(b.data.Download, t)
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(tl)
	return &core.Resource{URL: url, File: filepath.Base(url), Tag: tag}, nil
}

func matchLine(r []*regexp.Regexp, line string) []string {
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
	return results
}
