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

// Do will perform purge (or dryrun at least)
func Do(dir string, known, pinned []string, context cli.Settings) (bool, error) {
	if dir == "" {
		return false, errors.New("directory must be set")
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	var re []*regexp.Regexp
	for _, p := range pinned {
		r, err := regexp.Compile(p)
		if err != nil {
			return false, err
		}
		re = append(re, r)
	}
	found := false
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
			found = true
			context.LogCore("purging: %s\n", name)
			if !context.DryRun {
				if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
					return false, err
				}
			}
		}
	}
	return found, nil
}
