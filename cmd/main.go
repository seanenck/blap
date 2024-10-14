package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/seanenck/bd/internal/core"
	"github.com/seanenck/bd/internal/fetch"
)

const (
	check         = "check"
	upgrade       = "upgrade"
	configFileEnv = "BD_CONFIG_FILE"
)

func defaultConfig() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "bd", "config.yaml")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func help(msg string) error {
	if msg != "" {
		fmt.Printf("%s\n\n", msg)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe = filepath.Base(exe)
	fmt.Printf("%s\n", exe)
	fmt.Printf("  %-10s    check for updates\n", check)
	fmt.Printf("  %-10s    upgrade packages\n", upgrade)
	fmt.Println()
	fmt.Printf("configuration file: %s\n", defaultConfig())
	fmt.Printf("  (override using %s)\n", configFileEnv)
	return nil
}

func run() error {
	args := os.Args
	if len(args) != 2 {
		return help("invalid arguments, missing command")
	}
	input := os.Getenv(configFileEnv)
	if input == "" {
		input = defaultConfig()
	}
	dryRun := true
	cmd := args[1]
	switch cmd {
	case "help":
		return help("")
	case check:
	case upgrade:
		dryRun = false
	default:
		return help(fmt.Sprintf("unknown argument: %s", cmd))
	}
	if !core.PathExists(input) {
		return fmt.Errorf("config file does not exist: %s", input)
	}
	cfg, err := core.LoadConfig(input, dryRun)
	if err != nil {
		return err
	}
	return cfg.Process(fetch.ResourceFetcher{})
}
