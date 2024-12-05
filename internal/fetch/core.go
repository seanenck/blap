// Package fetch handles getting package source and data
package fetch

import (
	"errors"
	"iter"
	"net/http"
	"regexp"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/logging"
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
		Debug(logging.Category, string, ...any)
		ExecuteCommand(cmd string, args ...string) (string, error)
		Get(string) (*http.Response, error)
	}
	// Filterable is an interface to support arbitrary inputs that need to filter to tag sets
	Filterable interface {
		Get(Retriever, string) ([]byte, error)
		Arguments() []string
		NewLine(string) (string, error)
	}
	// Template are the parameters for templated items
	Template struct {
		Tag core.Version
	}
)

// CompileRegexp will allow compiling a regex with specific settings
func (ctx Context) CompileRegexp(re string, vals any) (*regexp.Regexp, error) {
	t, err := ctx.Templating(re, vals)
	if err != nil {
		return nil, err
	}
	return regexp.Compile(t)
}

// Templating handles common templating for various fetch strings
func (ctx Context) Templating(in string, vals any) (string, error) {
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
