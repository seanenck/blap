// Package cli handles CLI flag parsing
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

const (
	// ListCommand will list the active applications
	ListCommand CommandType = "list"
	// PurgeCommand is used to delete old artifacts/dirs/etc.
	PurgeCommand CommandType = "purge"
	// UpgradeCommand is used to update packages
	UpgradeCommand CommandType = "upgrade"
	// VersionCommand displays version information
	VersionCommand = "version"
	// CompletionsCommand generates completions
	CompletionsCommand = "completions"
	// CommitFlag confirms and therefore commits changes
	CommitFlag = "commit"
	// VerbosityFlag changes logging output
	VerbosityFlag = "verbosity"
	// ApplicationsFlag enables selected applications
	ApplicationsFlag = "filter-applications"
	// CleanDirFlag indicates directory cleanup should occur
	CleanDirFlag = "directories"
	// ReDeployFlag will indicate all apps should ignore the redeployment rules and force redeploy
	ReDeployFlag = "force-redeploy"
	// NegateFilter means to IGNORE filter applications
	NegateFilter            = "negate-filter"
	isFlag                  = "--"
	displayApplicationsFlag = isFlag + ApplicationsFlag
	displayVerbosityFlag    = isFlag + VerbosityFlag
	displayCommitFlag       = isFlag + CommitFlag
	displayCleanDirFlag     = isFlag + CleanDirFlag
	displayReDeployFlag     = isFlag + ReDeployFlag
	displayNegateFlag       = isFlag + NegateFilter
)

// CommandType are top-level commands
type CommandType string

// Parse will parse arguments to settings
func Parse(w io.Writer, t CommandType, args []string) (*Settings, error) {
	var appFilter string
	var negateFilter bool
	var cleanDirs bool
	var isReDeploy bool
	dryRun := true
	verbosity := InfoVerbosity
	if len(args) > 0 {
		set := flag.NewFlagSet("app", flag.ContinueOnError)
		verbose := set.Int(VerbosityFlag, InfoVerbosity, "set verbosity level")
		var apps *string
		var reDeploy *bool
		var dirs *bool
		var negate *bool
		var commit *bool
		switch t {
		case PurgeCommand:
			dirs = set.Bool(CleanDirFlag, false, "cleanup orphaned directories")
		case ListCommand, UpgradeCommand:
			apps = set.String(ApplicationsFlag, "", "filter processed applications")
			negate = set.Bool(NegateFilter, false, "negate application filter")
			if t == UpgradeCommand {
				reDeploy = set.Bool(ReDeployFlag, false, "redeploy all applications")
			}
		}
		needCommit := t == PurgeCommand || t == UpgradeCommand
		if needCommit {
			commit = set.Bool(CommitFlag, false, "confirm and commit changes")
		}
		if err := set.Parse(args); err != nil {
			return nil, err
		}
		verbosity = *verbose
		if verbosity < 0 {
			return nil, fmt.Errorf("verbosity must be >= 0 (%d)", verbosity)
		}
		switch t {
		case PurgeCommand:
			cleanDirs = *dirs
		case ListCommand, UpgradeCommand:
			appFilter = *apps
			negateFilter = *negate
			if reDeploy != nil {
				isReDeploy = *reDeploy
			}
			if negateFilter && len(appFilter) == 0 {
				return nil, errors.New("negate used without filters")
			}
		}
		if needCommit {
			dryRun = !*commit
			if dryRun && isReDeploy {
				return nil, errors.New("can not redeploy and dry-run")
			}
		}
	}
	ctx := &Settings{
		CleanDirs: cleanDirs,
		DryRun:    dryRun,
		Verbosity: verbosity,
		Purge:     t == PurgeCommand,
		Writer:    w,
		ReDeploy:  isReDeploy,
	}
	if err := ctx.CompileApplicationFilter(appFilter, negateFilter); err != nil {
		return nil, err
	}
	return ctx, nil
}
