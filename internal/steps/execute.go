// Package steps handles build step processing
package steps

import (
	"errors"
	"path/filepath"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/logging"
	"github.com/seanenck/blap/internal/util"
)

// Do will run each build step
func Do(steps []core.Step, builder util.Runner, ctx Context, e core.CommandEnv) error {
	if builder == nil {
		return errors.New("builder is unset")
	}
	if err := ctx.Valid(); err != nil {
		return err
	}
	environ := e.Variables.Set()
	defer environ.Unset()
	for _, step := range steps {
		for cmd := range step.Steps() {
			if len(cmd) == 0 {
				continue
			}
			to := ctx.Variables.Vars.Directories.Root
			if step.Directory != "" {
				sub, err := ctx.Variables.Template(step.Directory.String())
				if err != nil {
					return err
				}
				to = filepath.Join(to, sub)
			}
			clone := ctx.Variables.Vars.Clone()
			clone.Directories.Working = to
			v, err := core.NewValues(ctx.Variables.Name, clone)
			if err != nil {
				return err
			}
			template := func(in string) (string, error) {
				return v.Template(in)
			}
			exe, err := template(cmd[0].String())
			if err != nil {
				return err
			}
			var args []string
			for idx, a := range cmd {
				if idx == 0 {
					continue
				}
				res := a.String()
				t, err := template(res)
				if err != nil {
					return err
				}
				args = append(args, t)
			}
			if err := runStep(ctx, builder, to, exe, args, step.CommandEnv(), step.ClearEnv || e.Clear); err != nil {
				return err
			}
		}
	}
	return nil
}

func runStep(ctx Context, builder util.Runner, to, exe string, args []string, env core.CommandEnv, doClear bool) error {
	environ := env.Variables.Set()
	defer environ.Unset()
	run := util.RunSettings{}
	run.Dir = to
	if doClear {
		run.Env.Clear = true
	}
	ctx.Settings.LogDebug(logging.BuildCategory, "run: %v\n", run)
	ctx.Settings.LogDebug(logging.BuildCategory, "command: %s (%v)\n", exe, args)
	return builder.Run(run, exe, args...)
}
