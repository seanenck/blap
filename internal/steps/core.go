// Package steps are common process step definitions
package steps

import (
	"errors"
	"fmt"
	"path/filepath"

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
		Directories Directories
	}
	// Directories contain information about where the unpacked/working copy is stored
	Directories struct {
		Root    string
		Working string
		files   map[string]string
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
	n.Directories.files = v.Directories.files
	return n
}

// NewVariables will initialize the new variable for steps
func NewVariables() Variables {
	v := Variables{}
	v.Directories.files = make(map[string]string)
	return v
}

// Valid will check validity of the context
func (c Context) Valid() error {
	if c.Variables.Vars.Directories.Root == "" {
		return errors.New("directory not set")
	}
	if c.Variables.Vars.Directories.files == nil {
		return errors.New("files map not set")
	}
	return nil
}

// Version will get the tag as a version
func (v Variables) Version() core.Version {
	return core.Version(v.Tag)
}

// Installed is used to indicate that an object is 'installed'
func (d Directories) Installed() string {
	return d.newMarker("installed", "")
}

// GetFile will get a NewFile set file
func (v Variables) GetFile(name string) string {
	if v.Directories.files != nil {
		return v.Directories.files[name]
	}
	return ""
}

// NewFile will create a 'blap'-based file path and add it to the markers for further use
func (v Variables) NewFile(name string) string {
	path := v.Directories.newMarker(name, "data_")
	v.Directories.files[name] = path
	return path
}

func (d Directories) newMarker(name, sub string) string {
	return filepath.Join(d.Root, fmt.Sprintf(".blap_%s%s", sub, name))
}
