// Package processing handles the base configuration
package processing

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
)

type (
	// Configuration is the overall configuration
	Configuration struct {
		handler *processHandler
		context cli.Settings

		Indexing struct {
			Enabled bool `yaml:"enabled"`
			Strict  bool `yaml:"strict"`
		} `yaml:"indexing"`
		Directory       core.Resolved    `yaml:"directory"`
		Include         []core.Resolved  `yaml:"include"`
		Applications    core.AppSet      `yaml:"applications"`
		Parallelization int              `yaml:"parallelization"`
		Pinned          core.Pinned      `yaml:"pinned"`
		Connections     core.Connections `yaml:"connections"`
		Variables       core.Variables   `yaml:"variables"`
		Log             core.Resolved    `yaml:"logfile"`
		pinnedMatchers  []*regexp.Regexp
		logFile         string
		dir             string
	}
)

// NewFile will create a new directory-based file from the configuration
func (c Configuration) NewFile(file string) string {
	return filepath.Join(c.dir, file)
}

// IndexFile will get an index file to assist in managing operations
func (c Configuration) IndexFile(mode string) string {
	return c.NewFile(fmt.Sprintf(".blap.%s.index", mode))
}
