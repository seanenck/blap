// Package steps are common process step definitions
package steps

import (
	"errors"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
)

type (
	// Variables define step variables for command templating
	Variables struct {
		*core.Resource
		Directories struct {
			Root    string
			Working string
		}
	}
	// Context are step settings/context
	Context struct {
		Settings  cli.Settings
		Variables core.Values[Variables]
	}
)

// Clone will clone a variable set
func (v Variables) Clone() Variables {
	n := Variables{}
	n.Directories.Root = v.Directories.Root
	n.Directories.Working = v.Directories.Working
	n.Resource = v.Resource
	return n
}

// Valid will check validity of the context
func (c Context) Valid() error {
	if c.Variables.Vars.Resource == nil {
		return errors.New("no resource set")
	}
	if c.Variables.Vars.Directories.Root == "" {
		return errors.New("directory not set")
	}
	return nil
}
