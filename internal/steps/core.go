// Package steps are common process step definitions
package steps

import (
	"errors"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/env"
)

type (
	Variables struct {
		*asset.Resource
		Directory string
	}
	// Context are step settings/context
	Context struct {
		Settings  cli.Settings
		Variables env.Values[Variables]
	}
)

// Valid will check validity of the context
func (c Context) Valid() error {
	if c.Variables.Vars.Resource == nil {
		return errors.New("no resource set")
	}
	if c.Variables.Vars.Directory == "" {
		return errors.New("directory not set")
	}
	return nil
}
