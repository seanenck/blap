// Package main is the primary executable
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/seanenck/blap/internal/context"
	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/shell"
)

const configFileEnv = "BLAP_CONFIG_FILE"

var version = "development"

func defaultConfig() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "blap", "config.yaml")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func helpLine(flag, text string) {
	fmt.Printf("  %-15s %s\n", flag, text)
}

func help(msg string) error {
	if msg != "" {
		fmt.Printf("%s\n\n", msg)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", filepath.Base(exe))
	helpLine(shell.UpgradeCommand, "upgrade packages")
	helpLine(shell.VersionCommand, "display version information")
	helpLine(shell.PurgeCommand, "purge old versions")
	helpLine(shell.DisplayApplicationsFlag, "specify a subset of packages (regex)")
	helpLine(shell.DisplayDisableFlag, "disable applications (regex)")
	helpLine(shell.DisplayVerbosityFlag, "increase/decrease output verbosity")
	helpLine(shell.DisplayCommitFlag, "confirm and commit changes for actions")
	fmt.Println()
	fmt.Printf("configuration file: %s\n", defaultConfig())
	fmt.Printf("  (override using %s)\n", configFileEnv)
	fmt.Println()
	fmt.Printf("to handle github rate limiting, specify a token in configuration or via env\n")
	fmt.Printf("  %s (directly or as reference to a file)\n", strings.Join(fetch.TokenOptions, ","))
	return nil
}

func run() error {
	args := os.Args
	if len(args) < 2 {
		return help("invalid arguments, missing command")
	}
	input := os.Getenv(configFileEnv)
	if input == "" {
		input = defaultConfig()
	}
	purging := false
	cmd := args[1]
	switch cmd {
	case shell.CompletionsCommand:
		return shell.GenerateCompletions(os.Stdout)
	case "help":
		return help("")
	case shell.PurgeCommand:
		purging = true
	case shell.VersionCommand:
		fmt.Println(version)
		return nil
	case shell.UpgradeCommand:
	default:
		return help(fmt.Sprintf("unknown argument: %s", cmd))
	}

	var appFilter string
	var negateFilter bool
	dryRun := true
	verbosity := context.InfoVerbosity
	if len(args) > 2 {
		set := flag.NewFlagSet("app", flag.ExitOnError)
		var apps *string
		var disable *string
		if !purging {
			apps = set.String(shell.ApplicationsFlag, "", "limit application checks")
			disable = set.String(shell.DisableFlag, "", "disable applications")
		}
		verbose := set.Int(shell.VerbosityFlag, context.InfoVerbosity, "set verbosity level")
		commit := set.Bool(shell.CommitFlag, false, "confirm and commit changes")
		if err := set.Parse(args[2:]); err != nil {
			return err
		}
		verbosity = *verbose
		if verbosity < 0 {
			return help("verbosity must be >= 0")
		}
		if !purging {
			a := *apps
			d := *disable
			lengthApps := len(a)
			lengthDis := len(d)
			if lengthApps > 0 || lengthDis > 0 {
				if len(a) > 0 && len(d) > 0 {
					return help("can not limit applications and disable at the same time")
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
	if !core.PathExists(input) {
		return fmt.Errorf("config file does not exist: %s", input)
	}
	ctx := &context.Settings{
		DryRun:    dryRun,
		Verbosity: verbosity,
		Purge:     purging,
		Writer:    os.Stdout,
	}
	if err := ctx.CompileApplicationFilter(appFilter, negateFilter); err != nil {
		return err
	}
	cfg, err := core.LoadConfig(input, *ctx)
	if err != nil {
		return err
	}
	return cfg.Process(&fetch.ResourceFetcher{})
}
