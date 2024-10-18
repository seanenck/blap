// Package cli handles shell completions
package cli

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

// Completion is the shell completion template object
type Completion struct {
	Executable string
	Command    struct {
		Purge   string
		Upgrade string
	}
	Arg struct {
		Applications string
		Confirm      string
		Disable      string
		Include      string
	}
}

//go:embed shell/bash
var bashShell string

// GenerateCompletions will generate shell completions
func GenerateCompletions(w io.Writer) error {
	if w == nil {
		return nil
	}
	if shell := filepath.Base(os.Getenv("SHELL")); shell != "bash" {
		return fmt.Errorf("unable to generate completions for %s", shell)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	comp := Completion{}
	comp.Executable = filepath.Base(exe)
	comp.Command.Purge = PurgeCommand
	comp.Command.Upgrade = UpgradeCommand
	comp.Arg.Confirm = DisplayCommitFlag
	comp.Arg.Applications = DisplayApplicationsFlag
	comp.Arg.Disable = DisplayDisableFlag
	comp.Arg.Include = DisplayIncludeFlag
	t, err := template.New("sh").Parse(bashShell)
	if err != nil {
		return err
	}
	return t.Execute(w, comp)
}
