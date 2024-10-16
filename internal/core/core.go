// Package core handles the core configuration/processing
package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"github.com/seanenck/blap/internal/context"
	"github.com/seanenck/blap/internal/extract"
	"gopkg.in/yaml.v3"
)

type (
	processHandler struct {
		assets  []string
		updated []string
	}
	// Configuration is the overall configuration
	Configuration struct {
		context      context.Settings
		Token        string
		Directory    string
		Include      []string               `yaml:"include"`
		Applications map[string]Application `yaml:"applications"`
	}
	// GitHubMode indicates processing of a github project for upstreams
	GitHubMode struct {
		Project string `yaml:"project"`
		Asset   string `yaml:"asset"`
	}
	// TaggedMode means a repository+download is required to manage
	TaggedMode struct {
		Repository string   `yaml:"repository"`
		Download   string   `yaml:"download"`
		Filters    []string `yaml:"filters"`
	}
	// BranchMode will enable a repository+branch to pull a tarball
	BranchMode struct {
		Project string
		Branch  string
	}
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Priority   int              `yaml:"priority"`
		Disable    bool             `yaml:"disable"`
		GitHub     *GitHubMode      `yaml:"github"`
		Tagged     *TaggedMode      `yaml:"tagged"`
		Branch     *BranchMode      `yaml:"branch"`
		Extract    extract.Settings `yaml:"extract"`
		BuildSteps []struct {
			Directory string   `yaml:"directory"`
			Command   []string `yaml:"command"`
		} `yaml:"build"`
		Deploy []struct {
			Files       []string `yaml:"files"`
			Destination string   `yaml:"destination"`
		} `yaml:"deploy"`
	}

	// Fetcher provides the means to fetch application information
	Fetcher interface {
		SetContext(context.Settings)
		GitHub(GitHubMode) (*extract.Asset, error)
		Tagged(TaggedMode) (*extract.Asset, error)
		Download(bool, string, string) (bool, error)
		SetToken(string)
		Branch(BranchMode) (*extract.Asset, error)
	}
	appError struct {
		name string
		err  error
	}
)

func (a appError) Error() string {
	return fmt.Sprintf("application name: %s, error: %v", a.name, a.err)
}

func (c Configuration) resolveDir() string {
	return resolveDir(c.Directory)
}

func resolveDir(dir string) string {
	isHome := fmt.Sprintf("~%c", os.PathSeparator)
	if !strings.HasPrefix(dir, isHome) {
		return dir
	}
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return dir
	}
	return filepath.Join(h, strings.TrimPrefix(dir, isHome))
}

func isNil(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}

	return false
}

func (a Application) process(name string, c Configuration, fetcher Fetcher, handler interface {
	AddAsset(*extract.Asset)
	Acted(string)
},
) error {
	c.context.LogInfo("processing: %s\n", name)
	fetcher.SetToken(resolveDir(c.Token))
	cnt := 0
	for _, obj := range []interface{}{a.GitHub, a.Tagged, a.Branch} {
		if !isNil(obj) {
			cnt++
			if cnt > 1 {
				return fmt.Errorf("multiple modes enable, only one allowed: %v", obj)
			}
		}
	}
	var asset *extract.Asset
	var err error
	if a.GitHub != nil {
		asset, err = fetcher.GitHub(*a.GitHub)
	} else {
		if a.Tagged != nil {
			asset, err = fetcher.Tagged(*a.Tagged)
		} else {
			if a.Branch != nil {
				asset, err = fetcher.Branch(*a.Branch)
			} else {
				return fmt.Errorf("unknown mode: %v", a)
			}
		}
	}
	if err != nil {
		return err
	}
	if asset == nil {
		return fmt.Errorf("no asset return: %v", a)
	}
	if err := asset.SetAppData(name, c.resolveDir(), a.Extract, c.context); err != nil {
		return err
	}
	handler.AddAsset(asset)
	if c.context.Purge {
		return nil
	}

	did, err := fetcher.Download(c.context.DryRun, asset.URL(), asset.Archive())
	if err != nil {
		return err
	}
	if did {
		handler.Acted(name)
	}
	if c.context.DryRun {
		return nil
	}

	dest := asset.Unpack()
	if !PathExists(dest) {
		if err := asset.Extract(); err != nil {
			return err
		}
	}
	obj := struct {
		Tag string
	}{asset.Tag()}
	for _, step := range a.BuildSteps {
		cmd := step.Command
		if len(cmd) == 0 {
			continue
		}
		exe := resolveDir(cmd[0])
		var args []string
		for idx, a := range cmd {
			if idx == 0 {
				continue
			}
			res := resolveDir(a)
			t, err := template.New("t").Parse(res)
			if err != nil {
				return err
			}
			var b bytes.Buffer
			if err := t.Execute(&b, obj); err != nil {
				return err
			}
			args = append(args, b.String())
		}
		c := exec.Command(exe, args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		to := dest
		if step.Directory != "" {
			to = filepath.Join(to, step.Directory)
		}
		c.Dir = to
		if err := c.Run(); err != nil {
			return err
		}
	}
	for _, deploy := range a.Deploy {
		dir := resolveDir(deploy.Destination)
		for _, b := range deploy.Files {
			to := filepath.Join(dir, filepath.Base(b))
			src := filepath.Join(asset.Unpack(), b)
			if !PathExists(src) {
				return fmt.Errorf("unable to find source file: %s", src)
			}
			if PathExists(to) {
				if err := os.RemoveAll(to); err != nil {
					return err
				}
			}
			if err := os.Symlink(src, to); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *processHandler) Acted(name string) {
	h.updated = append(h.updated, name)
}

func (h *processHandler) AddAsset(a *extract.Asset) {
	for _, f := range []string{a.Unpack(), a.Archive()} {
		h.assets = append(h.assets, filepath.Base(f))
	}
}

// Process will process application definitions
func (c Configuration) Process(fetcher Fetcher) error {
	fetcher.SetContext(c.context)
	type apps struct {
		app  Application
		name string
	}
	var enabled []apps
	for name, app := range c.Applications {
		if app.Disable {
			continue
		}
		enabled = append(enabled, apps{app, name})
	}
	slices.SortFunc(enabled, func(left, right apps) int {
		return right.app.Priority - left.app.Priority
	})
	handler := &processHandler{}
	for _, a := range enabled {
		if err := a.app.process(a.name, c, fetcher, handler); err != nil {
			return appError{a.name, err}
		}
	}
	if c.context.Purge {
		dir := c.resolveDir()
		dirs, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, d := range dirs {
			name := d.Name()
			if !slices.Contains(handler.assets, name) {
				c.context.LogCore("purging: %s\n", name)
				if !c.context.DryRun {
					if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
						return err
					}
				}
			}
		}
	} else {
		for idx, update := range handler.updated {
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

func doDecode[T any](in string, o T) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(o); err != nil {
		return fmt.Errorf("file: %s -> %v", in, err)
	}
	return nil
}

// LoadConfig will initialize the configuration from a file
func LoadConfig(input string, context context.Settings) (Configuration, error) {
	c := Configuration{}
	c.context = context
	c.Applications = make(map[string]Application)
	if err := doDecode(input, &c); err != nil {
		return c, err
	}
	if len(c.Include) > 0 {
		var including []string
		for _, i := range c.Include {
			r := resolveDir(i)
			res := []string{r}
			c.context.LogDebug("including: %s\n", i)
			if strings.Contains(r, "*") {
				globbed, err := filepath.Glob(r)
				if err != nil {
					return c, err
				}
				res = globbed
			}
			including = append(including, res...)
		}
		for _, include := range including {
			c.context.LogDebug("loading included: %s\n", include)
			type included struct {
				Applications map[string]Application `yaml:"applications"`
				Disable      bool                   `yaml:"disable"`
			}
			var apps included
			if err := doDecode(include, &apps); err != nil {
				return c, err
			}
			if apps.Disable {
				continue
			}
			for k, v := range apps.Applications {
				if _, ok := c.Applications[k]; ok {
					return c, fmt.Errorf("%s is overwritten by config: %s", k, include)
				}
				c.Applications[k] = v
			}
		}
	}
	if context.FilterApplications() {
		sub := make(map[string]Application)
		for n, a := range c.Applications {
			if context.AllowApplication(n) {
				sub[n] = a
			}
		}
		c.Applications = sub
	}
	return c, nil
}

// PathExists indicates if a file exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
