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
		Files   map[string]string
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
	n.Directories.Files = v.Directories.Files
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

// Installed is used to indicate that an object is 'installed'
func (d Directories) Installed() string {
	return d.newMarker("installed", "")
}

// NewFile will create a 'blap'-based file path and add it to the markers for further use
func (d Directories) NewFile(name string) string {
	path := d.newMarker(name, "data_")
	d.Files[name] = path
	return path
}

func (d Directories) newMarker(name, sub string) string {
	return filepath.Join(d.Root, fmt.Sprintf(".blap_%s%s", sub, name))
}
