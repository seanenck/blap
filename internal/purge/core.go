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
func Do(dir string, known []string, context cli.Settings) error {
	if dir == "" {
		return errors.New("directory must be set")
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		name := d.Name()
		if !slices.Contains(known, name) {
			context.LogCore("purging: %s\n", name)
			if !context.DryRun {
				if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
