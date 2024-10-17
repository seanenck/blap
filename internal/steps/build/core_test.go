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
	"github.com/seanenck/blap/internal/util"
)

type mockRun struct {
	lastDir   string
	lastCmd   string
	lastArgs  []string
	lastEnv   []string
	lastClear bool
}

func (m *mockRun) Run(s util.RunSettings, c string, args ...string) error {
	m.lastDir = s.Dir
	m.lastArgs = args
	m.lastCmd = c
	m.lastEnv = s.Env.Values
	m.lastClear = s.Env.Clear
	return nil
}

func (m *mockRun) RunCommand(string, ...string) error {
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
	if err := build.Do([]types.Step{}, m, "", step, types.BuildEnvironment{}); err == nil || err.Error() != "destination must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := build.Do([]types.Step{}, m, "a", step, types.BuildEnvironment{}); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	step.Resource = env.Values[*asset.Resource]{Vars: &asset.Resource{}}
	if err := build.Do([]types.Step{}, nil, "a", step, types.BuildEnvironment{}); err == nil || err.Error() != "builder is unset" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	t.Setenv("HOME", "xyz")
	e, _ := env.NewValues("a", &asset.Resource{File: "A"})
	step.Resource = e
	if err := build.Do([]types.Step{{}, {Directory: "xyz", Command: []string{"~/exe", `~/{{ if eq $.Arch "fakearch" }}{{else}}{{ $.Vars.File }}{{end}}`}}}, m, "a", step, types.BuildEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a/xyz xyz/exe [xyz/A] [] false}" {
		t.Errorf("invalid result: %v", m)
	}
	e, _ = env.NewValues("A", &asset.Resource{})
	step.Resource = e
	if err := build.Do([]types.Step{{}, {Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a xyz/exe [xyz/A] [] false}" {
		t.Errorf("invalid result: %v", m)
	}
}

func TestEnv(t *testing.T) {
	m := &mockRun{}
	os.Clearenv()
	e, _ := env.NewValues("A", &asset.Resource{})
	step := steps.Context{}
	step.Settings = cli.Settings{}
	step.Resource = e
	if err := build.Do([]types.Step{{}, {Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := build.Do([]types.Step{{}, {Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{Clear: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	o := types.BuildEnvironment{}
	o.Clear = true
	if err := build.Do([]types.Step{{}, {Environment: o, Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := build.Do([]types.Step{{Environment: o}, {Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{Clear: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	o.Clear = false
	if err := build.Do([]types.Step{{}, {Environment: o, Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{Values: []string{"A"}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 1 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	o.Values = []string{"AA", "BB"}
	if err := build.Do([]types.Step{{}, {Environment: o, Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 2 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := build.Do([]types.Step{{}, {Environment: o, Command: []string{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.BuildEnvironment{Values: []string{"A"}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 3 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
}
