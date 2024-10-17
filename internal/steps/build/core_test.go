package build_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/env"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/steps/build"
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
	step := steps.Context{}
	step.Settings = cli.Settings{}
	step.Resource = env.Values[*asset.Resource]{}
	if err := build.Do([]types.Step{}, m, "", step); err == nil || err.Error() != "destination must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := build.Do([]types.Step{}, m, "a", step); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	step.Resource = env.Values[*asset.Resource]{Vars: &asset.Resource{}}
	if err := build.Do([]types.Step{}, nil, "a", step); err == nil || err.Error() != "builder is unset" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	t.Setenv("HOME", "xyz")
	e, _ := env.NewValues("a", &asset.Resource{File: "A"})
	step.Resource = e
	if err := build.Do([]types.Step{{}, {Directory: "xyz", Command: []string{"~/exe", `~/{{ if eq $.Arch "fakearch" }}{{else}}{{ $.Vars.File }}{{end}}`}}}, m, "a", step); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a/xyz xyz/exe [xyz/A]}" {
		t.Errorf("invalid result: %v", m)
	}
	e, _ = env.NewValues("A", &asset.Resource{})
	step.Resource = e
	if err := build.Do([]types.Step{{}, {Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a xyz/exe [xyz/A]}" {
		t.Errorf("invalid result: %v", m)
	}
}
