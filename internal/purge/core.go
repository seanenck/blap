// Package purge handles clearing out old variants
package purge

import (
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/seanenck/blap/internal/cli"
)

// Do will perform purge (or dryrun at least)
func Do(dir string, known []string, context cli.Settings) (bool, error) {
	if dir == "" {
		return false, errors.New("directory must be set")
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	found := false
	for _, d := range dirs {
		name := d.Name()
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
