// Package cli handles help output
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/seanenck/blap/internal/config/types"
)

func helpLine(w io.Writer, flag, text string) {
	fmt.Fprintf(w, "  %-15s %s\n", flag, text)
}

func configPath(root string) string {
	return filepath.Join(root, "blap", "config.yaml")
}

// Usage writes usage/help info
func Usage(w io.Writer) error {
	return help(w)
}

func help(w io.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s\n", filepath.Base(exe))
	helpLine(w, UpgradeCommand, "upgrade packages")
	helpLine(w, VersionCommand, "display version information")
	helpLine(w, PurgeCommand, "purge old versions")
	helpLine(w, DisplayApplicationsFlag, "specify a subset of packages (regex)")
	helpLine(w, DisplayDisableFlag, "disable applications (regex)")
	helpLine(w, DisplayIncludeFlag, "include specified files only (regex)")
	helpLine(w, DisplayVerbosityFlag, "increase/decrease output verbosity")
	helpLine(w, DisplayCommitFlag, "confirm and commit changes for actions")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "configuration file locations:")
	for _, c := range DefaultConfigs() {
		fmt.Fprintf(w, "- %s\n", c)
	}
	fmt.Fprintf(w, "(override using %s)\n", ConfigFileEnv)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "to handle github rate limiting, specify a token in configuration or via env:\n")
	fmt.Fprintf(w, "- %s\n", strings.Join(types.GitHubSettings{}.Env(), "\n- "))
	return nil
}
