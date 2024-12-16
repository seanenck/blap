package cli_test

import (
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
)

func TestParseDefaults(t *testing.T) {
	c, err := cli.Parse(nil, cli.PurgeCommand, []string{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c.Purge {
		t.Error("should have purged enabled")
	}
	if !c.DryRun {
		t.Error("should do dryrun by default")
	}
	if c.Verbosity != cli.InfoVerbosity {
		t.Errorf("should use info verbosity: %d", c.Verbosity)
	}
	if c.FilterApplications() {
		t.Error("filters should be off")
	}
	c, err = cli.Parse(nil, cli.UpgradeCommand, []string{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Purge {
		t.Error("should have purged disabled")
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := cli.Parse(nil, cli.PurgeCommand, []string{"-verbosity", "-1"}); err == nil || err.Error() != "verbosity must be >= 0 (-1)" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := cli.Parse(nil, cli.UpgradeCommand, []string{"-negate-filter"}); err == nil || err.Error() != "negate used without filters" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := cli.Parse(nil, cli.UpgradeCommand, []string{"-force-redeploy"}); err == nil || !strings.Contains(err.Error(), "not redeploy and dry-run") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestParse(t *testing.T) {
	c, err := cli.Parse(nil, cli.UpgradeCommand, []string{"-verbosity", "5", "-commit"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun {
		t.Errorf("invalid result: %v", c)
	}
	c, err = cli.Parse(nil, cli.UpgradeCommand, []string{"-verbosity", "5", "-commit", "-filter-applications=nvim", "--negate-filter"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun || !c.FilterApplications() || c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
	c, err = cli.Parse(nil, cli.UpgradeCommand, []string{"-verbosity", "5", "-commit", "-filter-applications=nvim"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun || !c.FilterApplications() || !c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
	c, err = cli.Parse(nil, cli.UpgradeCommand, []string{"-verbosity", "5", "-commit", "-filter-applications=nvim"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun || !c.FilterApplications() || !c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
	if c.CleanDirs {
		t.Error("invalid, cleandirs should be set")
	}
	c, err = cli.Parse(nil, cli.PurgeCommand, []string{"-directories", "-commit"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c.CleanDirs {
		t.Error("invalid, cleandirs should be set")
	}
	c, err = cli.Parse(nil, cli.ListCommand, []string{"-verbosity", "15", "-filter-applications=nvim", "--negate-filter"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 15 || !c.DryRun || !c.FilterApplications() || c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
	c, err = cli.Parse(nil, cli.ListCommand, []string{"-verbosity", "15", "-filter-applications=nvim"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 15 || !c.DryRun || !c.FilterApplications() || !c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
}
