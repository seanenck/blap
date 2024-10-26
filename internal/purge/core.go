// Package purge handles clearing out old variants
package purge

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/seanenck/blap/internal/cli"
)

// OnPurge is called when a purge is performed
type OnPurge func()

// Do will perform purge (or dryrun at least)
func Do(dir string, known, pinned []string, context cli.Settings, fxn OnPurge) error {
	if dir == "" {
		return errors.New("directory must be set")
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var re []*regexp.Regexp
	for _, p := range pinned {
		r, err := regexp.Compile(p)
		if err != nil {
			return err
		}
		re = append(re, r)
	}
	for _, d := range dirs {
		name := d.Name()
		pin := false
		for _, r := range re {
			if r.MatchString(name) {
				pin = true
				break
			}
		}
		if pin {
			continue
		}
		if !slices.Contains(known, name) {
			context.LogCore("purging: %s\n", name)
			fxn()
			if !context.DryRun {
				if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
