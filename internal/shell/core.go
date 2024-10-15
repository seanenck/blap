// Package shell handles shell/cli components
package shell

import (
	_ "embed"
	"os"
	"path/filepath"
	"text/template"
)

const (
	// PurgeCommand is used to delete old artifacts/dirs/etc.
	PurgeCommand = "purge"
	// UpgradeCommand is used to update packages
	UpgradeCommand = "upgrade"
	// VersionCommand displays version information
	VersionCommand = "version"
	// BashCommand generates bash completions
	BashCommand = "bash"
	// CommitFlag confirms and therefore commits changes
	CommitFlag = "confirm"
	// VerbosityFlag changes logging output
	VerbosityFlag = "verbosity"
	// ApplicationsFlag enables selected applications
	ApplicationsFlag = "applications"
	// DisableFlag disables selected applications
	DisableFlag = "disable"
	isFlag      = "-"
	// DisplayApplicationsFlag is the displayed version of application flag
	DisplayApplicationsFlag = isFlag + ApplicationsFlag
	// DisplayDisableFlag is the displayed version of disable flag
	DisplayDisableFlag = isFlag + DisableFlag
	// DisplayVerbosityFlag is the displayed version of verbosity flag
	DisplayVerbosityFlag = isFlag + VerbosityFlag
	// DisplayCommitFlag is the displayed version of the commit flag
	DisplayCommitFlag = isFlag + CommitFlag
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
	}
}

//go:embed bd.bash
var bashShell string

// BashCompletions generations bash completions
func BashCompletions() error {
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
	t, err := template.New("bash").Parse(bashShell)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, comp)
}
