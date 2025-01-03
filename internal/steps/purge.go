// Package steps handles clearing out old variants
package steps

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

// OnPurge is called when a purge is performed
type OnPurge func(string) bool

// Purge will perform purge operations (or dryrun at least)
func Purge(dir string, known []string, pinned []*regexp.Regexp, fxn OnPurge) error {
	if dir == "" {
		return errors.New("directory must be set")
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
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
			if fxn(name) {
				if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
