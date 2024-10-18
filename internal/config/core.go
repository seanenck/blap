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
		Token           string       `yaml:"token"`
		Directory       string       `yaml:"directory"`
		Include         []string     `yaml:"include"`
		Applications    types.AppSet `yaml:"applications"`
		Parallelization int          `yaml:"parallelization"`
		Pinned          types.Pinned `yaml:"pinned"`
	}
)
