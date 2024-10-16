// Package util handles commands
package util

import (
	"os"
	"os/exec"
)

type (
	// CommandRunner is the default command runner
	CommandRunner struct{}
	// Runner is the runner interface for exec'ing
	Runner interface {
		Run(string, ...string) error
		Output(string, ...string) ([]byte, error)
		RunIn(string, string, ...string) error
	}
)

// RunIn will run a command in a directory
func (r CommandRunner) RunIn(dest, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if dest != "" {
		c.Dir = dest
	}
	return c.Run()
}

// Run will run a command
func (r CommandRunner) Run(cmd string, args ...string) error {
	return r.RunIn("", cmd, args...)
}

// Output will get command output
func (r CommandRunner) Output(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}
