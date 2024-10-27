// Package steps handles clearing out old variants
package steps

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/seanenck/blap/internal/cli"
)

// OnPurge is called when a purge is performed
type OnPurge func(string)

// Purge will perform purge operations (or dryrun at least)
func Purge(dir string, known []string, pinned []*regexp.Regexp, context cli.Settings, fxn OnPurge) error {
	if dir == "" {
		return errors.New("directory must be set")
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	from := filepath.Base(dir)
	for _, d := range dirs {
		name := d.Name()
		pin := false
		for _, r := range pinned {
			if r.MatchString(name) {
				pin = true
				break
			}
		}
		if pin {
			continue
		}
		if !slices.Contains(known, name) {
			context.Purging(from, name)
			fxn(name)
			if !context.DryRun {
				if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
