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
			Enabled bool
			Strict  bool
		}
		Directory       core.Resolved
		Include         []core.Resolved
		Applications    core.AppSet
		Parallelization int
		Pinned          core.Pinned
		Connections     core.Connections
		Variables       core.Variables
		Logging         struct {
			File core.Resolved
			Size int64
		}
		pinnedMatchers []*regexp.Regexp
		logFile        string
		dir            string
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
