package steps_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/env"
	"github.com/seanenck/blap/internal/steps"
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
	m.lastEnv = os.Environ()
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
	if err := steps.Do([]types.Step{}, m, "", step, types.CommandEnvironment{}); err == nil || err.Error() != "destination must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := steps.Do([]types.Step{}, m, "a", step, types.CommandEnvironment{}); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	step.Resource = env.Values[*asset.Resource]{Vars: &asset.Resource{}}
	if err := steps.Do([]types.Step{}, nil, "a", step, types.CommandEnvironment{}); err == nil || err.Error() != "builder is unset" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	t.Setenv("HOME", "xyz")
	e, _ := env.NewValues("a", &asset.Resource{File: "A"})
	step.Resource = e
	if err := steps.Do([]types.Step{{}, {Directory: "xyz", Command: []types.Resolved{"~/exe", `~/{{ if eq $.Arch "fakearch" }}{{else}}{{ $.Vars.File }}{{end}}`}}}, m, "a", step, types.CommandEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a/xyz xyz/exe [xyz/A] [HOME=xyz] false}" {
		t.Errorf("invalid result: %v", m)
	}
	e, _ = env.NewValues("A", &asset.Resource{})
	step.Resource = e
	if err := steps.Do([]types.Step{{}, {Command: []types.Resolved{"~/$HOME/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.CommandEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a xyz/xyz/exe [xyz/A] [HOME=xyz] false}" {
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
	if err := steps.Do([]types.Step{{}, {Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.CommandEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := steps.Do([]types.Step{{}, {Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.CommandEnvironment{Clear: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	o := types.CommandEnvironment{}
	o.Clear = true
	if err := steps.Do([]types.Step{{}, {Environment: o, Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.CommandEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := steps.Do([]types.Step{{Environment: o}, {Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.CommandEnvironment{Clear: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	o.Clear = false
	o.Variables.Vars = make(map[string]types.Resolved)
	o.Variables.Vars["HOME"] = "1"
	if err := steps.Do([]types.Step{{}, {Environment: o, Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, types.CommandEnvironment{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 1 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	s := types.CommandEnvironment{}
	s.Variables.Vars = make(map[string]types.Resolved)
	s.Variables.Vars["HOME"] = "y"
	if err := steps.Do([]types.Step{{}, {Environment: o, Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, s); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 1 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	o.Variables.Vars["XYZ"] = "aaa"
	if err := steps.Do([]types.Step{{}, {Environment: o, Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, s); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 2 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	s.Variables.Vars["ZZZ"] = "aaz"
	if err := steps.Do([]types.Step{{}, {Environment: o, Command: []types.Resolved{"~/exe", "~/{{ $.Name }}"}}}, m, "a", step, s); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 3 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
}
