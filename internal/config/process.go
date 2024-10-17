// Package config handles processing yaml configs
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"sync"

	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/env"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/purge"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/steps/build"
	"github.com/seanenck/blap/internal/steps/deploy"
	"github.com/seanenck/blap/internal/util"
)

var processLock = &sync.Mutex{}

type (
	errorList      []error
	processHandler struct {
		assets  []string
		updated []string
	}
	// Executor is the process executor
	Executor interface {
		Do(Context) error
		Purge() (bool, error)
		Updated() []string
	}
	// Context allows processing an application (fetch, extract, build, deploy)
	Context struct {
		Name        string
		Application types.Application
		Fetcher     fetch.Retriever
		Runner      util.Runner
	}
)

func (e errorList) add(errs []chan error) errorList {
	for _, r := range errs {
		if err := <-r; err != nil {
			e = append(e, err)
		}
	}
	return e
}

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
	rsrc, err := ctx.Fetcher.Process(fetch.Context{Name: ctx.Name}, ctx.Application.Source.Items())
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
		processLock.Lock()
		c.handler.assets = append(c.handler.assets, filepath.Base(f))
		processLock.Unlock()
	}
	if c.context.Purge {
		return nil
	}

	did, err := ctx.Fetcher.Download(c.context.DryRun, rsrc.URL, rsrc.Paths.Archive)
	if err != nil {
		return err
	}
	if did {
		processLock.Lock()
		c.handler.updated = append(c.handler.updated, ctx.Name)
		processLock.Unlock()
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
	e, err := env.NewValues(ctx.Name, rsrc)
	if err != nil {
		return err
	}
	step := steps.Context{}
	step.Resource = e
	step.Settings = c.context
	if err := build.Do(ctx.Application.Build.Steps, ctx.Runner, dest, step, ctx.Application.Build.Environment); err != nil {
		return err
	}
	if err := deploy.Do(rsrc.Paths.Unpack, ctx.Application.Deploy.Artifacts, step); err != nil {
		return err
	}
	return nil
}

// Purge will run a purge operation
func (c Configuration) Purge() (bool, error) {
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
	if c.Parallelization < 0 {
		return fmt.Errorf("parallelization must be >= 0 (have: %d)", c.Parallelization)
	}
	fetcher.SetToken(c.context.Resolve(c.Token))
	var priorities []int
	prioritySet := make(map[int][]Context)
	for name, app := range c.Applications {
		has, ok := prioritySet[app.Priority]
		if !ok {
			has = []Context{}
			priorities = append(priorities, app.Priority)
		}
		has = append(has, Context{Name: name, Application: app, Fetcher: fetcher, Runner: runner})
		prioritySet[app.Priority] = has
	}
	sort.Ints(priorities)
	slices.Reverse(priorities)
	for _, p := range priorities {
		var errs []chan error
		var errorSet errorList
		for _, a := range prioritySet[p] {
			for len(errs) > c.Parallelization {
				errorSet = errorSet.add(errs)
				errs = []chan error{}
			}
			chn := make(chan error)
			go func(ctx Context, e chan error) {
				var cErr error
				if err := executor.Do(ctx); err != nil {
					cErr = fmt.Errorf("application '%s' error: %v", ctx.Name, err)
				}
				e <- cErr
			}(a, chn)
			errs = append(errs, chn)
		}
		errorSet = errorSet.add(errs)
		if len(errorSet) > 0 {
			return errors.Join(errorSet...)
		}
	}
	changed := false
	if c.context.Purge {
		purged, err := executor.Purge()
		if err != nil {
			return err
		}
		changed = purged
	} else {
		for idx, update := range executor.Updated() {
			if idx == 0 {
				changed = true
				c.context.LogCore("updates\n")
			}
			c.context.LogCore("  -> %s\n", update)
		}
	}
	if c.context.DryRun && changed {
		c.context.LogCore("\n[DRYRUN] impactful changes were not committed\n")
	}
	return nil
}
