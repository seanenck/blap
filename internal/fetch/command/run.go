// Package command can run an arbitrary command+args to get version information
package command

import (
	"errors"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/filtered"
)

type runFilterable struct {
	filtered.Base
	args []string
}

func (run runFilterable) Get(r fetch.Retriever, cmd string) ([]byte, error) {
	out, err := r.ExecuteCommand(cmd, run.args...)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func (run runFilterable) Arguments() []string {
	return run.args
}

func (run runFilterable) NewLine(line string) (string, error) {
	return line, nil
}

// Run will execute the given command+args
func Run(caller fetch.Retriever, ctx fetch.Context, a core.RunMode) (*core.Resource, error) {
	// check command before NewBase to get a more reasonable error message
	if a.Executable == "" {
		return nil, errors.New("command not set")
	}
	b, err := filtered.NewBase(a.Executable, a.Fetch, runFilterable{args: a.Arguments})
	if err != nil {
		return nil, err
	}
	return b.Get(caller, ctx)
}
