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
		URL         string
		Tag         string
		File        string
		Archive     string
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
	n.Tag = v.Tag
	n.URL = v.URL
	n.File = v.File
	n.Archive = v.Archive
	return n
}

// Valid will check validity of the context
func (c Context) Valid() error {
	if c.Variables.Vars.Directories.Root == "" {
		return errors.New("directory not set")
	}
	return nil
}

// Version will get the tag as a version
func (v Variables) Version() core.Version {
	return core.Version(v.Tag)
}
