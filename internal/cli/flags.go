// Package cli handles CLI flag parsing
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

const (
	// PurgeCommand is used to delete old artifacts/dirs/etc.
	PurgeCommand = "purge"
	// UpgradeCommand is used to update packages
	UpgradeCommand = "upgrade"
	// VersionCommand displays version information
	VersionCommand = "version"
	// CompletionsCommand generates completions
	CompletionsCommand = "completions"
	// CommitFlag confirms and therefore commits changes
	CommitFlag = "commit"
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

// Parse will parse arguments to settings
func Parse(w io.Writer, purging bool, args []string) (*Settings, error) {
	var appFilter string
	var negateFilter bool
	dryRun := true
	verbosity := InfoVerbosity
	if len(args) > 0 {
		set := flag.NewFlagSet("app", flag.ContinueOnError)
		var apps *string
		var disable *string
		if !purging {
			apps = set.String(ApplicationsFlag, "", "limit application checks")
			disable = set.String(DisableFlag, "", "disable applications")
		}
		verbose := set.Int(VerbosityFlag, InfoVerbosity, "set verbosity level")
		commit := set.Bool(CommitFlag, false, "confirm and commit changes")
		if err := set.Parse(args); err != nil {
			return nil, err
		}
		verbosity = *verbose
		if verbosity < 0 {
			return nil, fmt.Errorf("verbosity must be >= 0 (%d)", verbosity)
		}
		if !purging {
			a := *apps
			d := *disable
			lengthApps := len(a)
			lengthDis := len(d)
			if lengthApps > 0 || lengthDis > 0 {
				if len(a) > 0 && len(d) > 0 {
					return nil, errors.New("can not limit applications and disable at the same time")
				}
				if lengthApps > 0 {
					appFilter = a
				} else {
					negateFilter = true
					appFilter = d
				}
			}
		}
		dryRun = !*commit
	}
	ctx := &Settings{
		DryRun:    dryRun,
		Verbosity: verbosity,
		Purge:     purging,
		Writer:    w,
	}
	if err := ctx.CompileApplicationFilter(appFilter, negateFilter); err != nil {
		return nil, err
	}
	return ctx, nil
}
