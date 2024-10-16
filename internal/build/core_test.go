package build_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/build"
	"github.com/seanenck/blap/internal/cli"
)

type mockRun struct {
	lastDir  string
	lastCmd  string
	lastArgs []string
}

func (m *mockRun) RunIn(d, c string, args ...string) error {
	m.lastDir = d
	m.lastArgs = args
	m.lastCmd = c
	return nil
}

func (m *mockRun) Run(string, ...string) error {
	return nil
}

func (m *mockRun) Output(string, ...string) ([]byte, error) {
	return nil, nil
}

func TestDo(t *testing.T) {
	m := &mockRun{}
	if err := build.Do([]build.Step{}, m, "", nil, cli.Settings{}); err == nil || err.Error() != "destination must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := build.Do([]build.Step{}, m, "a", nil, cli.Settings{}); err == nil || err.Error() != "resource is unset" {
		t.Errorf("invalid error: %v", err)
	}
	if err := build.Do([]build.Step{}, nil, "a", 1, cli.Settings{}); err == nil || err.Error() != "builder is unset" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	t.Setenv("HOME", "xyz")
	if err := build.Do([]build.Step{{}, {Directory: "xyz", Command: []string{"~/exe", "~/{{ $.A }}"}}}, m, "a", struct{ A string }{"A"}, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a/xyz xyz/exe [xyz/A]}" {
		t.Errorf("invalid result: %v", m)
	}
	if err := build.Do([]build.Step{{}, {Command: []string{"~/exe", "~/{{ $.A }}"}}}, m, "a", struct{ A string }{"A"}, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a xyz/exe [xyz/A]}" {
		t.Errorf("invalid result: %v", m)
	}
}
