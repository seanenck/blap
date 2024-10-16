// Package config handles processing yaml configs
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/build"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/deploy"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/purge"
	"github.com/seanenck/blap/internal/util"
)

type (
	processHandler struct {
		assets  []string
		updated []string
	}
	// Configuration is the overall configuration
	Configuration struct {
		handler      *processHandler
		context      cli.Settings
		Token        string
		Directory    string
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
	// Executor is the process executor
	Executor interface {
		Do(Context) error
		Purge() error
		Updated() []string
	}
	// Context allows processing an application (fetch, extract, build, deploy)
	Context struct {
		Name        string
		Application Application
		Fetcher     fetch.Retriever
		Runner      util.Runner
	}
)

func (c Configuration) resolveDir() string {
	return c.context.Resolve(c.Directory)
}

// Do will perform processing of configuration components
func (c Configuration) Do(ctx Context) error {
	if ctx.Name == "" {
		return errors.New("name is required")
	}
	if ctx.Fetcher == nil || ctx.Runner == nil {
		return errors.New("fetcher and runner must be set")
	}
	if c.handler == nil {
		return errors.New("configuration not setup")
	}
	c.context.LogInfo("processing: %s\n", ctx.Name)
	ctx.Fetcher.SetToken(c.context.Resolve(c.Token))
	rsrc, err := ctx.Fetcher.Process(ctx.Fetcher, ctx.Application.GitHub, ctx.Application.Tagged)
	if err != nil {
		return err
	}
	if err := rsrc.SetAppData(ctx.Name, c.resolveDir(), ctx.Application.Extract); err != nil {
		return err
	}
	for _, f := range []string{rsrc.Paths.Unpack, rsrc.Paths.Archive} {
		if f == "" {
			continue
		}
		c.handler.assets = append(c.handler.assets, filepath.Base(f))
	}
	if c.context.Purge {
		return nil
	}

	did, err := ctx.Fetcher.Download(c.context.DryRun, rsrc.URL, rsrc.Paths.Archive)
	if err != nil {
		return err
	}
	if did {
		c.handler.updated = append(c.handler.updated, ctx.Name)
	}
	if c.context.DryRun {
		return nil
	}

	dest := rsrc.Paths.Unpack
	if !util.PathExists(dest) {
		c.context.LogDebug("extracting: %s\n", rsrc.File)
		if err := rsrc.Extract(ctx.Runner); err != nil {
			return err
		}
	}
	if err := build.Do(ctx.Application.BuildSteps, ctx.Runner, dest, rsrc, c.context); err != nil {
		return err
	}
	if err := deploy.Do(rsrc.Paths.Unpack, ctx.Application.Deploy, c.context); err != nil {
		return err
	}
	return nil
}

// Purge will run a purge operation
func (c Configuration) Purge() error {
	return purge.Do(c.resolveDir(), c.handler.assets, c.context)
}

// Updated gets the list of updated components
func (c Configuration) Updated() []string {
	if c.handler != nil {
		return c.handler.updated
	}
	return nil
}

// Process will process application definitions
func (c Configuration) Process(executor Executor, fetcher fetch.Retriever, runner util.Runner) error {
	if c.Applications == nil {
		return nil
	}
	if c.handler == nil {
		return errors.New("configuration not ready")
	}
	var enabled []Context
	for name, app := range c.Applications {
		enabled = append(enabled, Context{Name: name, Application: app, Fetcher: fetcher, Runner: runner})
	}
	slices.SortFunc(enabled, func(left, right Context) int {
		return right.Application.Priority - left.Application.Priority
	})
	for _, a := range enabled {
		if err := executor.Do(a); err != nil {
			return fmt.Errorf("application '%s' error: %v", a.Name, err)
		}
	}
	if c.context.Purge {
		if err := executor.Purge(); err != nil {
			return err
		}
	} else {
		for idx, update := range executor.Updated() {
			if idx == 0 {
				c.context.LogCore("updates\n")
			}
			c.context.LogCore("  -> %s\n", update)
		}
	}
	if c.context.DryRun {
		c.context.LogCore("\n[DRYRUN] impactful changes were not committed\n")
	}
	return nil
}
