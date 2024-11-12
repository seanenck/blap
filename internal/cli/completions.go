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
		CleanDirs    string
	}
}

var (
	//go:embed shell/bash
	bashShell string
	//go:embed shell/zsh
	zshShell string
)

// GenerateCompletions will generate shell completions
func GenerateCompletions(w io.Writer) error {
	if w == nil {
		return nil
	}
	var text string
	switch filepath.Base(os.Getenv("SHELL")) {
	case "bash":
		text = bashShell
	case "zsh":
		text = zshShell
	default:
		return fmt.Errorf("unable to generate completions for shell")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	comp := Completion{}
	comp.Executable = filepath.Base(exe)
	comp.Command.Purge = PurgeCommand
	comp.Command.Upgrade = UpgradeCommand
	comp.Arg.Confirm = displayCommitFlag
	comp.Arg.Applications = displayApplicationsFlag
	comp.Arg.Disable = displayDisableFlag
	comp.Arg.Include = displayIncludeFlag
	comp.Arg.CleanDirs = displayCleanDirFlag
	t, err := template.New("sh").Parse(text)
	if err != nil {
		return err
	}
	return t.Execute(w, comp)
}
