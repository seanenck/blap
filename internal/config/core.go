// Package config handles processing yaml configs
package config

import (
	"fmt"
	"path/filepath"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
)

type (
	// Configuration is the overall configuration
	Configuration struct {
		handler         *processHandler
		context         cli.Settings
		Indexing        bool              `yaml:"indexing"`
		Directory       types.Resolved    `yaml:"directory"`
		Include         []types.Resolved  `yaml:"include"`
		Applications    types.AppSet      `yaml:"applications"`
		Parallelization int               `yaml:"parallelization"`
		Pinned          types.Pinned      `yaml:"pinned"`
		Connections     types.Connections `yaml:"connections"`
		Variables       types.Variables   `yaml:"variables"`
	}
)

// IndexFile will get an index file to assist in managing operations
func (c Configuration) IndexFile(mode string) string {
	return filepath.Join(c.Directory.String(), fmt.Sprintf(".blap.%s.index", mode))
}
