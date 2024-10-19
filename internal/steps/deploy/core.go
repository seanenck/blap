// Package deploy handles artifact deployment
package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/util"
)

// Do will perform deployments from a source dir
func Do(src string, deploys []types.Artifact, ctx steps.Context) error {
	if len(deploys) == 0 {
		return nil
	}
	if src == "" {
		return errors.New("source directory required")
	}
	if err := ctx.Valid(); err != nil {
		return err
	}
	for _, deploy := range deploys {
		if deploy.Destination == "" {
			return errors.New("missing deploy destination")
		}
		dir := deploy.Destination.String()
		d, err := ctx.Resource.Template(dir)
		if err != nil {
			return err
		}
		dir = d
		for _, path := range deploy.Files {
			if path == "" {
				return errors.New("empty file not allowed")
			}
			b, err := ctx.Resource.Template(path.String())
			if err != nil {
				return err
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
