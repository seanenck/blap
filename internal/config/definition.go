// Package config handles processing yaml configs
package config

import (
	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/build"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/deploy"
	"github.com/seanenck/blap/internal/fetch"
)

type (
	// Configuration is the overall configuration
	Configuration struct {
		handler      *processHandler
		context      cli.Settings
		Token        string                 `yaml:"token"`
		Directory    string                 `yaml:"directory"`
		Include      []string               `yaml:"include"`
		Applications map[string]Application `yaml:"applications"`
	}
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Priority   int               `yaml:"priority"`
		Disable    bool              `yaml:"disable"`
		GitHub     *fetch.GitHubMode `yaml:"github"`
		Tagged     *fetch.TaggedMode `yaml:"tagged"`
		Extract    asset.Settings    `yaml:"extract"`
		BuildSteps []build.Step      `yaml:"build"`
		Deploy     []deploy.Artifact `yaml:"deploy"`
	}
)
