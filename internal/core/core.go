// Package core handles the core configuration/processing
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seanenck/bd/internal/extract"
	"gopkg.in/yaml.v3"
)

type (
	// Configuration is the overall configuration
	Configuration struct {
		dryRun       bool
		Directory    string
		Applications map[string]Application `yaml:"applications"`
	}
	// Remote is the remote/upstream information for an application
	Remote struct {
		Upstream string   `yaml:"upstream"`
		Download string   `yaml:"download"`
		Filters  []string `yaml:"filters"`
		Asset    string   `yaml:"asset"`
	}
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Disable bool   `yaml:"disable"`
		Mode    string `yaml:"mode"`
		Remote  Remote `yaml:"remote"`
		Extract struct {
			Command []string `yaml:"command"`
		} `yaml:"extract"`
		Binaries struct {
			Files       []string `yaml:"files"`
			Destination string   `yaml:"destination"`
		} `yaml:"binaries"`
	}

	// Fetcher provides the means to fetch application information
	Fetcher interface {
		GitHub(Remote) (*extract.Asset, error)
		Tagged(Remote) (*extract.Asset, error)
		Download(bool, string, string) (bool, error)
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
	var asset *extract.Asset
	var err error
	switch a.Mode {
	case "github":
		asset, err = fetcher.GitHub(a.Remote)
	case "tagged":
		asset, err = fetcher.Tagged(a.Remote)
	default:
		return false, fmt.Errorf("unknown mode for binary handling: %s", a.Mode)
	}
	if err != nil {
		return false, err
	}
	if asset == nil {
		return false, fmt.Errorf("no asset found: %v", a)
	}
	if err := asset.SetAppData(name, resolveDir(c.Directory), a.Extract.Command); err != nil {
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
			fmt.Printf("skipping: %s\n", name)
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
	for idx, update := range updated {
		if idx == 0 {
			fmt.Println("updates found")
		}
		fmt.Printf("  -> %s\n", update)
	}
	return nil
}

// LoadConfig will initialize the configuration from a file
func LoadConfig(input string, dryRun bool) (Configuration, error) {
	c := Configuration{}
	c.dryRun = dryRun
	data, err := os.ReadFile(input)
	if err != nil {
		return c, err
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, err
	}
	return c, nil
}

// PathExists indicates if a file exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
