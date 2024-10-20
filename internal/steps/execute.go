// Package steps handles build step processing
package steps

import (
	"errors"
	"path/filepath"

	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/util"
)

// Do will run each build step
func Do(steps []types.Step, builder util.Runner, ctx Context, env types.CommandEnvironment) error {
	if builder == nil {
		return errors.New("builder is unset")
	}
	if err := ctx.Valid(); err != nil {
		return err
	}
	template := func(in string) (string, error) {
		return ctx.Variables.Template(in)
	}
	env.Variables.Set()
	defer env.Variables.Unset()
	for _, step := range steps {
		cmd := step.Command
		if len(cmd) == 0 {
			continue
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
		to := ctx.Variables.Vars.Directory
		if step.Directory != "" {
			sub, err := template(step.Directory.String())
			if err != nil {
				return err
			}
			to = filepath.Join(to, sub)
		}
		if err := runStep(ctx, builder, to, exe, args, step.Environment, step.Environment.Clear || env.Clear); err != nil {
			return err
		}
	}
	return nil
}

func runStep(ctx Context, builder util.Runner, to, exe string, args []string, env types.CommandEnvironment, doClear bool) error {
	env.Variables.Set()
	defer env.Variables.Unset()
	run := util.RunSettings{}
	run.Dir = to
	if doClear {
		run.Env.Clear = true
	}
	ctx.Settings.LogDebug("run: %v\n", run)
	ctx.Settings.LogDebug("command: %s (%v)\n", exe, args)
	return builder.Run(run, exe, args...)
}
