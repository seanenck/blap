package steps_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/blap/internal/steps"
)

func TestValid(t *testing.T) {
	c := steps.Context{}
	if err := c.Valid(); err == nil || err.Error() != "directory not set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables.Vars.Directories.Root = "abc"
	if err := c.Valid(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestClone(t *testing.T) {
	v := steps.Variables{}
	v.Directories.Root = "xyz"
	v.File = "a"
	v.URL = "xy"
	v.Archive = "id"
	v.Tag = "v1.2.3"
	v.Directories.Working = "work"
	v.Directories.Files = map[string]string{
		"xyz": "111",
	}
	n := v.Clone()
	if n.Directories.Root != "xyz" || n.Directories.Working != "work" || n.File != "a" || n.URL != "xy" || n.Archive != "id" || n.Tag != "v1.2.3" || n.Version().Full() != "1.2.3" || fmt.Sprintf("%v", n.Directories.Files) != "map[xyz:111]" {
		t.Errorf("invalid clone: %v (%s, %v)", n, n.Version(), n.Directories.Files)
	}
}

func TestFiles(t *testing.T) {
	v := steps.Variables{}
	v.Directories.Root = "xyz"
	v.Directories.Files = map[string]string{}
	if p := v.Directories.NewFile("vars"); p != "xyz/.blap_data_vars" {
		t.Error("invalid marker")
	}
	if v.Directories.Installed() != "xyz/.blap_installed" || v.Directories.Files["vars"] != "xyz/.blap_data_vars" {
		t.Error("invalid markers")
	}
}
