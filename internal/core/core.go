// Package core handles the core configuration/processing
package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/seanenck/bd/internal/extract"
	"gopkg.in/yaml.v3"
)

type (
	// Configuration is the overall configuration
	Configuration struct {
		dryRun       bool
		Token        string
		Directory    string
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
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Disable bool        `yaml:"disable"`
		GitHub  *GitHubMode `yaml:"github"`
		Tagged  *TaggedMode `yaml:"tagged"`
		Extract struct {
			NoDepth bool     `yaml:"nodepth"`
			Command []string `yaml:"command"`
		} `yaml:"extract"`
		Binaries struct {
			Files       []string `yaml:"files"`
			Destination string   `yaml:"destination"`
		} `yaml:"binaries"`
	}

	// Fetcher provides the means to fetch application information
	Fetcher interface {
		GitHub(GitHubMode) (*extract.Asset, error)
		Tagged(TaggedMode) (*extract.Asset, error)
		Download(bool, string, string) (bool, error)
		SetToken(string)
	}
	appError struct {
		name string
		err  error
	}
)

func (a appError) Error() string {
	return fmt.Sprintf("application name: %s, error: %v", a.name, a.err)
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

func (a Application) process(name string, c Configuration, fetcher Fetcher) (bool, error) {
	fmt.Printf("processing: %s\n", name)
	fetcher.SetToken(resolveDir(c.Token))
	if a.GitHub != nil && a.Tagged != nil {
		return false, fmt.Errorf("multiple modes enable, only one allowed: %v", a)
	}
	var asset *extract.Asset
	var err error
	if a.GitHub != nil {
		asset, err = fetcher.GitHub(*a.GitHub)
	} else {
		if a.Tagged != nil {
			asset, err = fetcher.Tagged(*a.Tagged)
		} else {
			return false, fmt.Errorf("unknown mode: %v", a)
		}
	}
	if err != nil {
		return false, err
	}
	if asset == nil {
		return false, fmt.Errorf("no asset found: %v", a)
	}
	if err := asset.SetAppData(name, resolveDir(c.Directory), !a.Extract.NoDepth, a.Extract.Command); err != nil {
		return false, err
	}

	did, err := fetcher.Download(c.dryRun, asset.URL(), asset.Archive())
	if err != nil {
		return false, err
	}
	if c.dryRun {
		return did, nil
	}

	dest := asset.Unpack()
	if !PathExists(dest) {
		if err := asset.Extract(); err != nil {
			return false, err
		}
	}
	dir := resolveDir(a.Binaries.Destination)
	for _, b := range a.Binaries.Files {
		to := filepath.Join(dir, filepath.Base(b))
		src := filepath.Join(asset.Unpack(), b)
		if !PathExists(src) {
			return false, fmt.Errorf("unable to find binary: %s", src)
		}
		if PathExists(to) {
			if err := os.Remove(to); err != nil {
				return false, err
			}
		}
		if err := os.Symlink(src, to); err != nil {
			return false, err
		}
	}
	return did, nil
}

// Process will process application definitions
func (c Configuration) Process(fetcher Fetcher) error {
	var updated []string
	for name, app := range c.Applications {
		if app.Disable {
			continue
		}
		did, err := app.process(name, c, fetcher)
		if err != nil {
			return appError{name, err}
		}
		if did {
			updated = append(updated, name)
		}
	}
	text := "found"
	if !c.dryRun {
		text = "applied"
	}
	for idx, update := range updated {
		if idx == 0 {
			fmt.Printf("\nupdates %s\n", text)
		}
		fmt.Printf("  -> %s\n", update)
	}
	return nil
}

// LoadConfig will initialize the configuration from a file
func LoadConfig(input string, dryRun bool, apps, disable []string) (Configuration, error) {
	c := Configuration{}
	c.dryRun = dryRun
	data, err := os.ReadFile(input)
	if err != nil {
		return c, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&c); err != nil {
		return c, err
	}
	isDisable := len(disable) > 0
	if len(apps) > 0 || isDisable {
		sub := make(map[string]Application)
		for n, a := range c.Applications {
			allow := false
			if isDisable {
				allow = !slices.Contains(disable, n)
			} else {
				if slices.Contains(apps, n) {
					allow = true
				}
			}
			if allow {
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
