// Package deploy handles artifact deployment
package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/util"
)

type (
	// Artifact are definitions of what to deploy
	Artifact struct {
		Files       []string `yaml:"files"`
		Destination string   `yaml:"destination"`
	}
)

// Do will perform deployments from a source dir
func Do(src string, deploys []Artifact, ctx cli.Settings) error {
	if len(deploys) == 0 {
		return nil
	}
	if src == "" {
		return errors.New("source directory required")
	}
	for _, deploy := range deploys {
		if deploy.Destination == "" {
			return errors.New("missing deploy destination")
		}
		dir := ctx.Resolve(deploy.Destination)
		for _, b := range deploy.Files {
			if b == "" {
				return errors.New("empty file not allowed")
			}
			to := filepath.Join(dir, filepath.Base(b))
			src := filepath.Join(src, b)
			if !util.PathExists(src) {
				return fmt.Errorf("unable to find source file: %s", src)
			}
			os.RemoveAll(to)
			if err := os.Symlink(src, to); err != nil {
				return err
			}
		}
	}
	return nil
}
