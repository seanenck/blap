package cli_test

import (
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
)

func TestParseDefaults(t *testing.T) {
	c, err := cli.Parse(nil, true, []string{})
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
	c, err = cli.Parse(nil, false, []string{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Purge {
		t.Error("should have purged disabled")
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := cli.Parse(nil, true, []string{"-verbosity", "-1"}); err == nil || err.Error() != "verbosity must be >= 0 (-1)" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := cli.Parse(nil, false, []string{"-applications", "1", "-disable", "2", "-verbosity", "1"}); err == nil || err.Error() != "can not limit applications and disable at the same time" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := cli.Parse(nil, false, []string{"-applications", "-1"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := cli.Parse(nil, true, []string{"-applications", "-1"}); err == nil || !strings.Contains(err.Error(), "flag provided but not defined:") {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := cli.Parse(nil, false, []string{"-force-redeploy"}); err == nil || !strings.Contains(err.Error(), "not redeploy and dry-run") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestParse(t *testing.T) {
	c, err := cli.Parse(nil, false, []string{"-verbosity", "5", "-commit"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun {
		t.Errorf("invalid result: %v", c)
	}
	c, err = cli.Parse(nil, false, []string{"-verbosity", "5", "-commit", "-applications=nvim"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun || !c.FilterApplications() || !c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
	c, err = cli.Parse(nil, false, []string{"-verbosity", "5", "-commit", "-disable=nvim"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.Verbosity != 5 || c.DryRun || !c.FilterApplications() || c.AllowApplication("nvim") {
		t.Errorf("invalid result: %v", c)
	}
	if c.CleanDirs {
		t.Error("invalid, cleandirs should be set")
	}
	c, err = cli.Parse(nil, true, []string{"-directories"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c.CleanDirs {
		t.Error("invalid, cleandirs should be set")
	}
}
