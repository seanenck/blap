// Package cli handles help output
package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/seanenck/blap/internal/core"
)

const exe = "blap"

func helpLine(w io.Writer, sub bool, flag, text string) {
	spacing := ""
	if sub {
		spacing = "  "
	}
	fmt.Fprintf(w, "  %-15s %s\n", fmt.Sprintf("%s%s", spacing, flag), text)
}

func configPath(root string) string {
	return filepath.Join(root, "blap", "config.toml")
}

// Usage writes usage/help info
func Usage(w io.Writer) error {
	return help(w)
}

func help(w io.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	fmt.Fprintf(w, "%s\n", exe)
	helpLine(w, false, VersionCommand, "display version information")
	helpLine(w, false, ListCommand, "list managed package set")
	helpLine(w, false, UpgradeCommand, "upgrade packages")
	helpLine(w, true, displayApplicationsFlag, "specify a subset of packages (regex)")
	helpLine(w, true, displayDisableFlag, "disable applications (regex)")
	helpLine(w, true, displayIncludeFlag, "include specified files only (regex)")
	helpLine(w, true, displayReDeployFlag, "redeploy all packages (ignoring application flags)")
	helpLine(w, false, PurgeCommand, "purge old versions")
	helpLine(w, true, displayCleanDirFlag, "cleanup orphan directories during purge")
	helpLine(w, false, displayVerbosityFlag, "increase/decrease output verbosity")
	helpLine(w, false, displayCommitFlag, "confirm and commit changes for actions")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "configuration file locations:")
	for _, c := range DefaultConfigs() {
		fmt.Fprintf(w, "- %s\n", c)
	}
	fmt.Fprintf(w, "(override using %s)\n", ConfigFileEnv)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "to handle github rate limiting, specify a token in configuration or via env:\n")
	fmt.Fprintf(w, "- %s\n", strings.Join(core.GitHubSettings{}.Env(), "\n- "))
	return nil
}
