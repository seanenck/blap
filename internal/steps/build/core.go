// Package build handles build step processing
package build

import (
	"errors"
	"path/filepath"

	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/util"
)

// Do will run each build step
func Do(steps []types.Step, builder util.Runner, destination string, ctx steps.Context, env types.BuildEnvironment) error {
	if destination == "" {
		return errors.New("destination must be set")
	}
	if builder == nil {
		return errors.New("builder is unset")
	}
	if err := ctx.Valid(); err != nil {
		return err
	}
	for _, step := range steps {
		cmd := step.Command
		if len(cmd) == 0 {
			continue
		}
		exe := cmd[0].String()
		var args []string
		for idx, a := range cmd {
			if idx == 0 {
				continue
			}
			res := a.String()
			t, err := ctx.Resource.Template(res)
			if err != nil {
				return err
			}
			args = append(args, t)
		}
		to := destination
		if step.Directory != "" {
			to = filepath.Join(to, step.Directory.String())
		}
		run := util.RunSettings{}
		run.Dir = to
		run.Env.Values = env.Values
		run.Env.Values = append(run.Env.Values, step.Environment.Values...)
		if step.Environment.Clear || env.Clear {
			run.Env.Clear = true
		}
		if err := builder.Run(run, exe, args...); err != nil {
			return err
		}
	}
	return nil
}
