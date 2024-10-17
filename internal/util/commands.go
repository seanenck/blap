// Package util handles commands
package util

import (
	"os"
	"os/exec"
)

type (
	// CommandRunner is the default command runner
	CommandRunner struct{}
	// RunSettings configure how a command is run
	RunSettings struct {
		Dir string
		Env struct {
			Clear  bool
			Values []string
		}
	}
	// Runner is the runner interface for exec'ing
	Runner interface {
		RunCommand(string, ...string) error
		Output(string, ...string) ([]byte, error)
		Run(RunSettings, string, ...string) error
	}
)

// Run will run a command with settings
func (r CommandRunner) Run(settings RunSettings, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if settings.Dir != "" {
		c.Dir = settings.Dir
	}
	if settings.Env.Clear || len(settings.Env.Values) > 0 {
		env := os.Environ()
		if settings.Env.Clear {
			env = []string{}
		}
		env = append(env, settings.Env.Values...)
		c.Env = env
	}
	return c.Run()
}

// RunCommand will run a command with default settings
func (r CommandRunner) RunCommand(cmd string, args ...string) error {
	return r.Run(RunSettings{}, cmd, args...)
}

// Output will get command output
func (r CommandRunner) Output(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}
