// Package cli handles shell completions
package cli

import (
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
			List    string
		}
		Params struct {
			Upgrade string
			Purge   string
			List    string
		}
		Arg struct {
			Applications string
			ForceDeploy  string
			Negate       string
			Confirm      string
			CleanDirs    string
		}
	}
)

//go:embed shell/*
var files embed.FS

func readFile(file string) ([]byte, error) {
	return files.ReadFile(filepath.Join(completionsDir, file))
}

// GenerateCompletions will generate shell completions
func GenerateCompletions(w io.Writer) error {
	if w == nil {
		return nil
	}
	comp := Completion{}
	comp.Executable = exe
	comp.Command.List = string(ListCommand)
	comp.Command.Purge = string(PurgeCommand)
	comp.Command.Upgrade = string(UpgradeCommand)
	comp.Arg.Confirm = displayCommitFlag
	comp.Arg.Applications = displayApplicationsFlag
	comp.Arg.CleanDirs = displayCleanDirFlag
	comp.Arg.ForceDeploy = displayReDeployFlag
	comp.Arg.Negate = displayNegateFlag

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
	comp.Params.Purge = strings.Join([]string{comp.Arg.Confirm, comp.Arg.CleanDirs}, " ")
	comp.Params.Upgrade = strings.Join([]string{comp.Arg.Confirm, comp.Arg.Applications, comp.Arg.Negate, comp.Arg.ForceDeploy}, " ")
	comp.Params.List = strings.Join([]string{comp.Arg.Applications, comp.Arg.Negate}, " ")
	t, err := template.New("sh").Parse(string(text))
	if err != nil {
		return err
	}
	return t.Execute(w, comp)
}
