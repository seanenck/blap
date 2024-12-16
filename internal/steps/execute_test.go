package steps_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
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
	if err := steps.Do([]core.Step{}, nil, step, core.CommandEnv{}); err == nil || err.Error() != "builder is unset" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	t.Setenv("HOME", "xyz")
	vars := steps.NewVariables()
	vars.Directories.Root = "a"
	vars.File = "A"
	e, _ := core.NewValues("xyz", vars)
	step.Variables = e
	if err := steps.Do([]core.Step{{}, {Directory: "{{ $.Name }}", Command: []interface{}{"~/exe/{{ $.Vars.Directories.Working }}", `~/{{ if eq $.Arch "fakearch" }}{{else}}{{ $.Vars.File }}{{end}}`}}}, m, step, core.CommandEnv{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a/xyz xyz/exe/a/xyz [xyz/A] [HOME=xyz] false}" {
		t.Errorf("invalid result: %v", m)
	}
	vars.File = "A"
	e, _ = core.NewValues("A", vars)
	step.Variables = e
	if err := steps.Do([]core.Step{{}, {Command: []interface{}{"~/$HOME/exe", "~/{{ $.Name }}"}}}, m, step, core.CommandEnv{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m) != "&{a xyz/xyz/exe [xyz/A] [HOME=xyz] false}" {
		t.Errorf("invalid result: %v", m)
	}
}

func TestEnv(t *testing.T) {
	m := &mockRun{}
	os.Clearenv()
	vars := steps.NewVariables()
	vars.Directories.Root = "a"
	vars.File = "A"
	e, _ := core.NewValues("a", vars)
	step := steps.Context{}
	step.Settings = cli.Settings{}
	step.Variables = e
	if err := steps.Do([]core.Step{{}, {Command: []interface{}{"~/exe", "~/{{ $.Name }}"}}}, m, step, core.CommandEnv{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := steps.Do([]core.Step{{}, {Command: []interface{}{[]interface{}{"~/exe"}, []interface{}{"~/{{ $.Name }}"}}}}, m, step, core.CommandEnv{Clear: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := steps.Do([]core.Step{{}, {ClearEnv: true, Command: []interface{}{"~/exe", "~/{{ $.Name }}"}}}, m, step, core.CommandEnv{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if err := steps.Do([]core.Step{{ClearEnv: true}, {Command: []interface{}{"~/exe/{{ $.Name }}", "~/{{ $.Name }}"}}}, m, step, core.CommandEnv{Clear: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.lastClear || len(m.lastEnv) > 0 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	if m.lastCmd != "~/exe/a" {
		t.Errorf("invalid cmd: %s", m.lastCmd)
	}
	v := core.Variables{}
	v = append(v, struct {
		Key   string
		Value core.Resolved
	}{"HOME", "1"})
	if err := steps.Do([]core.Step{{}, {Variables: v, Command: []interface{}{"~/exe", "~/{{ $.Name }}"}}}, m, step, core.CommandEnv{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 1 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	s := core.CommandEnv{}
	v = core.Variables{}
	v = append(v, struct {
		Key   string
		Value core.Resolved
	}{"HOME", "y"})
	if err := steps.Do([]core.Step{{}, {Variables: v, Command: []interface{}{"~/exe", "~/{{ $.Name }}"}}}, m, step, s); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 1 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	v = append(v, struct {
		Key   string
		Value core.Resolved
	}{"XYZ", "aaa"})
	if err := steps.Do([]core.Step{{}, {Variables: v, Command: []interface{}{"~/exe", "~/{{ $.Name }}"}}}, m, step, s); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 2 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
	v = append(v, struct {
		Key   string
		Value core.Resolved
	}{"ZZZ", "aaz"})
	if err := steps.Do([]core.Step{{}, {Variables: v, Command: []interface{}{"~/exe", "~/{{ $.Name }}"}}}, m, step, s); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.lastClear || len(m.lastEnv) != 3 {
		t.Errorf("invalid env: %v %v", m.lastClear, m.lastEnv)
	}
}
