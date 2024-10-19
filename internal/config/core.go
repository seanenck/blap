// Package config handles processing yaml configs
package config

import (
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
)

type (
	// Configuration is the overall configuration
	Configuration struct {
		handler         *processHandler
		context         cli.Settings
		Directory       types.Resolved    `yaml:"directory"`
		Include         []types.Resolved  `yaml:"include"`
		Applications    types.AppSet      `yaml:"applications"`
		Parallelization int               `yaml:"parallelization"`
		Pinned          types.Pinned      `yaml:"pinned"`
		Connections     types.Connections `yaml:"connections"`
	}
)
