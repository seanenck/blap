package steps_test

import (
	"testing"

	"github.com/seanenck/blap/internal/steps"
)

func TestValid(t *testing.T) {
	c := steps.Context{}
	if err := c.Valid(); err == nil || err.Error() != "directory not set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables.Vars.Directories.Root = "abc"
	if err := c.Valid(); err == nil || err.Error() != "files map not set" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestClone(t *testing.T) {
	v := steps.NewVariables()
	v.Directories.Root = "xyz"
	v.File = "a"
	v.URL = "xy"
	v.Archive = "id"
	v.Tag = "v1.2.3"
	v.Directories.Working = "work"
	v.NewFile("111")
	n := v.Clone()
	if n.Directories.Root != "xyz" || n.Directories.Working != "work" || n.File != "a" || n.URL != "xy" || n.Archive != "id" || n.Tag != "v1.2.3" || n.Version().Full() != "1.2.3" || n.GetFile("111") != "xyz/.blap_data_111" {
		t.Errorf("invalid clone: %v (%s, %s)", n, n.Version(), n.GetFile("111"))
	}
}

func TestFiles(t *testing.T) {
	v := steps.NewVariables()
	v.Directories.Root = "xyz"
	if p := v.NewFile("vars"); p != "xyz/.blap_data_vars" {
		t.Error("invalid marker")
	}
	if v.Directories.Installed() != "xyz/.blap_installed" || v.GetFile("vars") != "xyz/.blap_data_vars" {
		t.Error("invalid markers")
	}
	v = steps.Variables{}
	if p := v.GetFile("xyz"); p != "" {
		t.Error("should be empty string")
	}
}
