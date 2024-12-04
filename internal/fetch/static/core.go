// Package static handles static definitions of resources
package static

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
)

// New creates a new static resource
func New(ctx fetch.Context, a core.StaticMode) (*core.Resource, error) {
	if a.URL == "" {
		return nil, errors.New("upstream URL not set")
	}
	if a.Tag == "" {
		return nil, errors.New("tag required for static mode")
	}
	upstream, err := ctx.Templating(a.URL.String(), struct{ Tag string }{a.Tag})
	if err != nil {
		return nil, err
	}
	file := a.File
	if file == "" {
		file = fmt.Sprintf("%s-%s", a.Tag, filepath.Base(upstream))
	} else {
		tl, err := ctx.Templating(file, &fetch.Template{Tag: fetch.Version(a.Tag)})
		if err != nil {
			return nil, err
		}
		file = tl
	}
	return &core.Resource{URL: upstream, Tag: a.Tag, File: file}, nil
}
