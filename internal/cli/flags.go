// Package cli handles CLI flag parsing
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"regexp"
)

const (
	// ListCommand will list the active applications
	ListCommand = "list"
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
	// IncludeFlag allows for filtering included files
	IncludeFlag = "include"
	// CleanDirFlag indicates directory cleanup should occur
	CleanDirFlag            = "directories"
	isFlag                  = "--"
	displayIncludeFlag      = isFlag + IncludeFlag
	displayApplicationsFlag = isFlag + ApplicationsFlag
	displayDisableFlag      = isFlag + DisableFlag
	displayVerbosityFlag    = isFlag + VerbosityFlag
	displayCommitFlag       = isFlag + CommitFlag
	displayCleanDirFlag     = isFlag + CleanDirFlag
)

// Parse will parse arguments to settings
func Parse(w io.Writer, purging bool, args []string) (*Settings, error) {
	var appFilter string
	var negateFilter bool
	var includeFilter string
	var cleanDirs bool
	dryRun := true
	verbosity := InfoVerbosity
	if len(args) > 0 {
		set := flag.NewFlagSet("app", flag.ContinueOnError)
		var apps *string
		var disable *string
		var include *string
		var dirs *bool
		if purging {
			dirs = set.Bool(CleanDirFlag, false, "cleanup orphaned directories")
		} else {
			apps = set.String(ApplicationsFlag, "", "limit application checks")
			disable = set.String(DisableFlag, "", "disable applications")
			include = set.String(IncludeFlag, "", "include only matched files")
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
		if purging {
			cleanDirs = *dirs
		} else {
			a := *apps
			d := *disable
			includeFilter = *include
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
	var includeReg *regexp.Regexp
	if includeFilter != "" {
		re, err := regexp.Compile(includeFilter)
		if err != nil {
			return nil, err
		}
		includeReg = re
	}
	ctx := &Settings{
		CleanDirs: cleanDirs,
		DryRun:    dryRun,
		Verbosity: verbosity,
		Purge:     purging,
		Writer:    w,
		Include:   includeReg,
	}
	if err := ctx.CompileApplicationFilter(appFilter, negateFilter); err != nil {
		return nil, err
	}
	return ctx, nil
}
