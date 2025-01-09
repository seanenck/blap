package steps_test

import (
	"errors"
	"testing"

	"github.com/seanenck/blap/internal/steps"
)

type mockFetch struct {
	err error
}

func (m *mockFetch) Download(bool, string, string) (bool, error) {
	return false, m.err
}

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
	v := steps.NewVariables(&mockFetch{})
	v.Directories.Root = "xyz"
	v.File = "a"
	v.URL = "xy"
	v.Archive = "id"
	v.Tag = "v1.2.3"
	v.Directories.Working = "work"
	v.GetFile("111")
	n := v.Clone()
	if n.Directories.Root != "xyz" || n.Directories.Working != "work" || n.File != "a" || n.URL != "xy" || n.Archive != "id" || n.Tag != "v1.2.3" || n.Version().Full() != "1.2.3" || n.GetFile("111") != "xyz/.blap_data_111" {
		t.Errorf("invalid clone: %v (%s, %s)", n, n.Version(), n.GetFile("111"))
	}
	n.Download("", "")
}

func TestFiles(t *testing.T) {
	v := steps.NewVariables(nil)
	v.Directories.Root = "xyz"
	if p := v.GetFile("vars"); p != "xyz/.blap_data_vars" {
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

func TestDownload(t *testing.T) {
	m := &mockFetch{}
	s := steps.NewVariables(m)
	if code, err := s.Download("", ""); err != nil || code != 0 {
		t.Errorf("invalid code/err: %d %v", code, err)
	}
	m.err = errors.New("error")
	s = steps.NewVariables(m)
	if code, err := s.Download("", ""); err == nil || err.Error() != "error" || code != 1 {
		t.Errorf("invalid code/err: %d %v", code, err)
	}
}
