// Package processing handles processing yaml configs
package processing

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/util"
)

var processLock = &sync.Mutex{}

type (
	errorList []error
	// Change are update/purge change sets
	Change struct {
		Name    string
		Details string
	}
	processHandler struct {
		changed []Change
	}
	// Executor is the process executor
	Executor interface {
		Do(Context) error
		Purge(string, []string, steps.OnPurge) error
		Changed() []Change
	}
	// Context allows processing an application (fetch, extract, command)
	Context struct {
		Name        string
		Application core.Application
		Fetcher     fetch.Retriever
		Runner      util.Runner
		Executor    Executor
	}
	// Index is the persisted index state information
	Index struct {
		Names []string `json:"names,omitempty"`
		Dirs  []string `json:"dirs,omitempty"`
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

// Do will perform processing of configuration components
func (c Configuration) Do(ctx Context) error {
	if ctx.Name == "" {
		return errors.New("name is required")
	}
	if ctx.Fetcher == nil || ctx.Runner == nil || ctx.Executor == nil {
		return errors.New("fetcher, runner, and executor must be set")
	}
	if c.handler == nil {
		return errors.New("configuration not setup")
	}
	c.context.LogDebug("processing: %s\n", ctx.Name)
	rsrc, err := ctx.Fetcher.Process(fetch.Context{Name: ctx.Name}, ctx.Application.Source.Items())
	if err != nil {
		return err
	}
	if rsrc == nil {
		return errors.New("unexpected nil resource")
	}
	to := filepath.Join(c.Directory.String(), ctx.Name)
	hasDest := util.PathExists(to)
	if !hasDest {
		if c.context.Purge {
			return nil
		}
		if !c.context.DryRun {
			c.context.LogDebug("making subdirectory: %s\n", to)
			if err := os.Mkdir(to, 0o755); err != nil {
				return err
			}
		}
	}
	if err := rsrc.SetAppData(ctx.Name, to, ctx.Application.Extract); err != nil {
		return err
	}
	knownAssets := []string{}
	for _, f := range []string{rsrc.Paths.Unpack, rsrc.Paths.Archive} {
		if f == "" {
			continue
		}
		knownAssets = append(knownAssets, filepath.Base(f))
	}
	onChange := func(detail string) {
		obj := Change{Name: ctx.Name, Details: detail}
		processLock.Lock()
		if !slices.ContainsFunc(c.handler.changed, func(c Change) bool {
			return obj.Name == c.Name
		}) {
			c.handler.changed = append(c.handler.changed, obj)
		}
		processLock.Unlock()
	}
	if c.context.Purge {
		return ctx.Executor.Purge(to, knownAssets, onChange)
	}

	did, err := ctx.Fetcher.Download(c.context.DryRun, rsrc.URL, rsrc.Paths.Archive)
	if err != nil {
		return err
	}
	if did {
		onChange(rsrc.Tag)
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
	vars := steps.Variables{}
	vars.Resource = rsrc
	vars.Directories.Root = dest
	e, err := core.NewValues(ctx.Name, vars)
	if err != nil {
		return err
	}
	step := steps.Context{}
	step.Variables = e
	step.Settings = c.context
	processLock.Lock()
	defer processLock.Unlock()
	if err := steps.Do(ctx.Application.Commands.Steps, ctx.Runner, step, ctx.Application.Commands.Environment); err != nil {
		return err
	}
	return nil
}

// Purge will run purge on inputs
func (c Configuration) Purge(dir string, assets []string, fxn steps.OnPurge) error {
	return steps.Purge(dir, assets, c.pinnedMatchers, c.context, fxn)
}

// Changed gets the list of changed components
func (c Configuration) Changed() []Change {
	return c.handler.changed
}

func (c Configuration) cleanDirectories(restrict []string) ([]string, error) {
	dir := c.Directory.String()
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	restricted := len(restrict) > 0
	var results []string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		name := d.Name()
		if _, ok := c.Applications[name]; ok {
			continue
		}
		if restricted {
			if !slices.Contains(restrict, name) {
				continue
			}
		}
		matched := false
		for _, p := range c.pinnedMatchers {
			if p.MatchString(name) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		results = append(results, name)
		c.context.LogCore("removing directory: %s\n", name)
		if c.context.DryRun {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return nil, err
		}
	}
	return results, nil
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
	if !c.Indexing.Enabled && c.Indexing.Strict {
		return errors.New("can not enable strict indexing without indexing enabled")
	}
	mode := "update"
	if c.context.Purge {
		mode = "purge"
	}
	indexFile := c.IndexFile(mode)
	idx := Index{}
	if c.Indexing.Enabled {
		if !c.context.DryRun {
			exists := util.PathExists(indexFile)
			if c.Indexing.Strict && !exists {
				return fmt.Errorf("index not found: %s (strict mode)", indexFile)
			}
			if exists {
				b, err := os.ReadFile(indexFile)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(b, &idx); err != nil {
					return err
				}
			}
		}
		c.context.LogDebug("index: %v", idx)
	}
	hasIndex := len(idx.Names) > 0
	fetcher.SetConnections(c.Connections)
	var priorities []int
	prioritySet := make(map[int][]Context)
	for name, app := range c.Applications {
		if hasIndex {
			if !slices.Contains(idx.Names, name) {
				continue
			}
		}
		has, ok := prioritySet[app.Priority]
		if !ok {
			has = []Context{}
			priorities = append(priorities, app.Priority)
		}
		has = append(has, Context{Name: name, Application: app, Fetcher: fetcher, Runner: runner, Executor: executor})
		prioritySet[app.Priority] = has
	}
	sort.Ints(priorities)
	slices.Reverse(priorities)
	c.Variables.Set()
	defer c.Variables.Unset()
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
	changed := executor.Changed()
	isDryRun := false
	newIndex := Index{}
	if c.context.Purge {
		if c.context.CleanDirs {
			dirs := idx.Dirs
			if c.context.DryRun {
				dirs = []string{}
			}
			results, err := c.cleanDirectories(dirs)
			if err != nil {
				return err
			}
			if c.context.DryRun && len(results) > 0 {
				newIndex.Dirs = results
				isDryRun = true
			}
		}
	} else {
		for _, update := range changed {
			c.context.Updating(update.Name, update.Details)
		}
	}
	if c.context.DryRun {
		if len(changed) > 0 {
			isDryRun = true
			for _, o := range changed {
				newIndex.Names = append(newIndex.Names, o.Name)
			}
		}
	}
	if c.Indexing.Enabled {
		removeIndex := util.PathExists(indexFile)
		if c.context.DryRun && (len(newIndex.Dirs) > 0 || len(newIndex.Names) > 0) {
			removeIndex = false
			b, err := json.Marshal(newIndex)
			if err != nil {
				return err
			}
			if err := os.WriteFile(indexFile, b, 0o644); err != nil {
				return err
			}
		}
		if removeIndex {
			return os.Remove(indexFile)
		}
	}
	if isDryRun {
		c.context.LogCore("\n[DRYRUN] impactful changes were not committed\n")
	}
	return nil
}
