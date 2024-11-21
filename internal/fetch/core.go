// Package fetch handles getting package source and data
package fetch

import (
	"errors"
	"fmt"
	"iter"
	"net/http"
	"regexp"
	"strings"

	"github.com/seanenck/blap/internal/core"
)

type (
	// Context is passed to processing to handle various inputs/values
	Context struct {
		Name string
	}
	// Backend to override calling conventions to external sources
	Backend interface {
		Do(*http.Request) (*http.Response, error)
		Output(string, ...string) ([]byte, error)
	}
	// Retriever provides the means to fetch application information
	Retriever interface {
		Download(bool, string, string) (bool, error)
		SetConnections(core.Connections)
		Process(Context, iter.Seq[any]) (*core.Resource, error)
		GitHubFetch(ownerRepo, call string, to any) error
		Debug(string, ...any)
		ExecuteCommand(cmd string, args ...string) (string, error)
		Get(string) (*http.Response, error)
	}
	// Filterable is an interface to support arbitrary inputs that need to filter to tag sets
	Filterable interface {
		Get(Retriever, string) ([]byte, error)
		Match([]*regexp.Regexp, string) ([]string, error)
	}
	// Template are the parameters for templated items
	Template struct {
		Tag Version
	}
	// Version is a tag that could be a semantic version
	Version string
)

// Major is the major component
func (v Version) Major() string {
	m, _, _, _ := v.parse()
	return m
}

// Minor is the minor component
func (v Version) Minor() string {
	_, m, _, _ := v.parse()
	return m
}

// Patch is the patch component
func (v Version) Patch() string {
	_, _, p, _ := v.parse()
	return p
}

// Remainder is anything after patch
func (v Version) Remainder() string {
	_, _, _, r := v.parse()
	return r
}

// Version is the major.minor.patch.remainder (- prefix 'v')
func (v Version) Version() string {
	major, minor, patch, remainder := v.parse()
	if major == "" {
		return ""
	}
	if minor == "" {
		return major
	}
	if patch == "" {
		return fmt.Sprintf("%s.%s", major, minor)
	}
	if remainder == "" {
		return fmt.Sprintf("%s.%s.%s", major, minor, patch)
	}
	return strings.Join([]string{major, minor, patch, remainder}, ".")
}

func (v Version) parse() (string, string, string, string) {
	parts := strings.Split(string(v), ".")
	major := strings.TrimPrefix(parts[0], "v")
	var minor, patch, left string
	if len(parts) > 1 {
		minor = parts[1]
		if len(parts) > 2 {
			patch = parts[2]
			if len(parts) > 3 {
				left = strings.Join(parts[3:], ".")
			}
		}
	}
	return major, minor, patch, left
}

// CompileRegexp will allow compiling a regex with specific settings
func (ctx Context) CompileRegexp(re string, vals *Template) (*regexp.Regexp, error) {
	t, err := ctx.Templating(re, vals)
	if err != nil {
		return nil, err
	}
	return regexp.Compile(t)
}

// Templating handles common templating for various fetch strings
func (ctx Context) Templating(in string, vals *Template) (string, error) {
	if ctx.Name == "" {
		return "", errors.New("context missing name")
	}
	v, err := core.NewValues(ctx.Name, vals)
	if err != nil {
		return "", err
	}
	b, err := v.Template(in)
	if err != nil {
		return "", err
	}
	return b, nil
}
