// Package build handles build step processing
package build

import (
	"bytes"
	"errors"
	"path/filepath"
	"text/template"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/util"
)

type (
	// Step is a build process step
	Step struct {
		Directory string   `yaml:"directory"`
		Command   []string `yaml:"command"`
	}
)

// Do will run each build step
func Do(steps []Step, builder util.Runner, destination string, rsrc any, context cli.Settings) error {
	if destination == "" {
		return errors.New("destination must be set")
	}
	if rsrc == nil {
		return errors.New("resource is unset")
	}
	if builder == nil {
		return errors.New("builder is unset")
	}
	for _, step := range steps {
		cmd := step.Command
		if len(cmd) == 0 {
			continue
		}
		exe := context.Resolve(cmd[0])
		var args []string
		for idx, a := range cmd {
			if idx == 0 {
				continue
			}
			res := context.Resolve(a)
			t, err := template.New("t").Parse(res)
			if err != nil {
				return err
			}
			var b bytes.Buffer
			if err := t.Execute(&b, rsrc); err != nil {
				return err
			}
			args = append(args, b.String())
		}
		to := destination
		if step.Directory != "" {
			to = filepath.Join(to, step.Directory)
		}
		if err := builder.RunIn(to, exe, args...); err != nil {
			return err
		}
	}
	return nil
}
