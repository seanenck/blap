// Package cli handles shell completions
package cli

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const completionsDir = "shell"

type (
	// Completion is the shell completion template object
	Completion struct {
		Executable string
		Command    struct {
			Purge   string
			Upgrade string
		}
		Params struct {
			Upgrade CompletionCommand
			Purge   CompletionCommand
		}
		Arg struct {
			Applications string
			Confirm      string
			Disable      string
			Include      string
			CleanDirs    string
		}
	}

	// CompletionCommand are specifics for completing a command
	CompletionCommand struct {
		Main string
		Sub  string
	}
)

//go:embed shell/*
var files embed.FS

func readFile(file string) ([]byte, error) {
	return files.ReadFile(filepath.Join(completionsDir, file))
}

func (c Completion) newParam(command string) (string, error) {
	b, err := readFile(fmt.Sprintf("params.%s.sh", strings.ToLower(command)))
	if err != nil {
		return "", err
	}
	t, err := template.New("t").Parse(string(b))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateCompletions will generate shell completions
func GenerateCompletions(w io.Writer) error {
	if w == nil {
		return nil
	}
	exe, err := baseExe()
	if err != nil {
		return err
	}
	comp := Completion{}
	comp.Executable = exe
	comp.Command.Purge = PurgeCommand
	comp.Command.Upgrade = UpgradeCommand
	comp.Arg.Confirm = displayCommitFlag
	comp.Arg.Applications = displayApplicationsFlag
	comp.Arg.Disable = displayDisableFlag
	comp.Arg.Include = displayIncludeFlag
	comp.Arg.CleanDirs = displayCleanDirFlag

	file := filepath.Base(os.Getenv("SHELL"))
	switch file {
	case "bash":
	case "zsh":
	default:
		return fmt.Errorf("unable to generate completions for shell")
	}
	text, err := readFile(fmt.Sprintf("completions.%s", file))
	if err != nil {
		return err
	}
	up, err := comp.newParam(UpgradeCommand)
	if err != nil {
		return err
	}
	purge, err := comp.newParam(PurgeCommand)
	if err != nil {
		return err
	}
	comp.Params.Purge.Main = strings.Join([]string{comp.Arg.Confirm, comp.Arg.CleanDirs}, " ")
	comp.Params.Purge.Sub = purge
	comp.Params.Upgrade.Main = strings.Join([]string{comp.Arg.Confirm, comp.Arg.Applications, comp.Arg.Disable, comp.Arg.Include}, " ")
	comp.Params.Upgrade.Sub = up
	t, err := template.New("sh").Parse(string(text))
	if err != nil {
		return err
	}
	return t.Execute(w, comp)
}
