package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seanenck/bd/internal/extract"
)

type (
	Configuration struct {
		dryRun       bool
		Directory    string
		Applications []Application
	}
	Remote struct {
		Upstream string
		Download string
		Filters  []string
		Asset    string
	}
	Application struct {
		Name     string
		Mode     string
		Remote   Remote
		Binaries struct {
			Files       []string
			Destination string
		}
	}

	Fetcher interface {
		GitHub(Remote) (*extract.Asset, error)
		Tagged(Remote) (*extract.Asset, error)
		Download(bool, string, string) (bool, error)
	}
	appError struct {
		index int
		err   error
	}
)

func (a appError) Error() string {
	return fmt.Sprintf("application index: %d, error: %v", a.index, a.err)
}

func (c Configuration) ResolveDirectory() string {
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

func (a Application) process(c Configuration, fetcher Fetcher) (bool, error) {
	if a.Name == "" {
		return false, fmt.Errorf("no name set: %v", a)
	}
	fmt.Printf("processing: %s\n", a.Name)
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
	asset.SetAppData(a.Name, c.ResolveDirectory())
	if !asset.HasExtractor() {
		return false, fmt.Errorf("no asset extractor: %s", asset.Archive())
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

func (c Configuration) Process(fetcher Fetcher) error {
	var updated []string
	for idx, app := range c.Applications {
		did, err := app.process(c, fetcher)
		if err != nil {
			return appError{idx, err}
		}
		if did {
			updated = append(updated, app.Name)
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

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
