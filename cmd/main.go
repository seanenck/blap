// Package main is the primary executable
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/fetch/retriever"
	"github.com/seanenck/blap/internal/processing"
	"github.com/seanenck/blap/internal/util"
)

var version = "development"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args
	if len(args) < 2 {
		return errors.New("invalid arguments, missing command")
	}
	var commandType cli.CommandType
	cmd := args[1]
	switch cmd {
	case cli.CompletionsCommand:
		return cli.GenerateCompletions(os.Stdout)
	case "help":
		return cli.Usage(os.Stdout)
	case string(cli.PurgeCommand):
		commandType = cli.PurgeCommand
	case string(cli.ListCommand):
		commandType = cli.ListCommand
	case cli.VersionCommand:
		fmt.Println(version)
		return nil
	case string(cli.UpgradeCommand):
		commandType = cli.UpgradeCommand
	default:
		return fmt.Errorf("unknown argument: %s", cmd)
	}

	sub := []string{}
	if len(args) > 1 {
		sub = args[2:]
	}
	ctx, err := cli.Parse(os.Stdout, commandType, sub)
	if err != nil {
		return err
	}
	input := os.Getenv(cli.ConfigFileEnv)
	if input == "" {
		for _, c := range cli.DefaultConfigs() {
			if util.PathExists(c) {
				input = c
			}
		}
	}
	if input == "" || !util.PathExists(input) {
		return fmt.Errorf("config file not set or does not exist: %s", input)
	}
	cfg, err := processing.Load(input, *ctx)
	if err != nil {
		return err
	}
	if commandType == cli.ListCommand {
		return cfg.List(os.Stdout)
	}
	return cfg.Process(cfg, &retriever.ResourceFetcher{Context: *ctx}, util.CommandRunner{})
}
