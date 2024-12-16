package core_test

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/seanenck/blap/internal/core"
)

func TestSourceItems(t *testing.T) {
	s := core.Application{}
	cnt := 0
	for range s.Items() {
		cnt++
	}
	if cnt != 5 {
		t.Errorf("invalid reflection count %d", cnt)
	}
}

func TestResolve(t *testing.T) {
	os.Clearenv()
	v := core.Resolved("")
	if v.String() != "" {
		t.Errorf("invalid result: %v", v)
	}
	t.Setenv("HOME", "abc")
	v = core.Resolved("~/")
	if v.String() != "abc" {
		t.Errorf("invalid result: %v", v)
	}
	t.Setenv("XXX", "123")
	v = core.Resolved("~/$XXX")
	if v.String() != "abc/123" {
		t.Errorf("invalid result: %v", v)
	}
	v = core.Resolved("~/$XXX{{ $.Abc }}")
	if v.String() != "abc/123{{ $.Abc }}" {
		t.Errorf("invalid result: %v", v)
	}
	v = core.Resolved("~/$XXX{{ if ne $.Config.Arch \"111\" }}$XXX{{end}}")
	if v.String() != "abc/123123" {
		t.Errorf("invalid result: %v", v)
	}
}

func TestGitHubToken(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	token := core.GitHubSettings{}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if token.Value() != nil {
		t.Errorf("invalid value: %s", token.Value())
	}
	token = core.GitHubSettings{Token: "xyz"}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if fmt.Sprintf("%v", token.Value()) != "[xyz]" {
		t.Errorf("invalid value: %s", token.Value())
	}
	t.Setenv("HOME", "zzz")
	token = core.GitHubSettings{Token: []interface{}{"$HOME/xyz", "$HOME/111"}}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if fmt.Sprintf("%v", token.Value()) != "[zzz/xyz zzz/111]" {
		t.Errorf("invalid value: %s", token.Value())
	}
}

func TestVarSetUnset(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	var val core.Variables
	obj := val.Set()
	obj.Unset()
	val = core.Variables{}
	obj = val.Set()
	obj.Unset()
	os.Setenv("HOME", "1")
	os.Setenv("A_TEST", "0")
	val = append(val, struct {
		Key   string
		Value core.Resolved
	}{"A_TEST", "~/2"})
	val = append(val, struct {
		Key   string
		Value core.Resolved
	}{"THIS_IS_A_TEST", "3"})
	obj = val.Set()
	if os.Getenv("A_TEST") != "1/2" || os.Getenv("THIS_IS_A_TEST") != "3" {
		t.Errorf("invalid env")
	}
	obj.Unset()
	if os.Getenv("A_TEST") != "0" || os.Getenv("THIS_IS_A_TEST") != "" {
		t.Errorf("invalid env: %s", os.Getenv("THIS_IS_A_TEST"))
	}
}

func TestCommandEnv(t *testing.T) {
	a := core.Application{}
	a.ClearEnv = true
	a.Variables = append(a.Variables, struct {
		Key   string
		Value core.Resolved
	}{})
	e := a.CommandEnv()
	if !e.Clear || len(e.Variables) != 1 {
		t.Error("invalid conversion")
	}
	s := core.Application{}
	s.ClearEnv = true
	s.Variables = append(s.Variables, struct {
		Key   string
		Value core.Resolved
	}{})
	s.Variables = append(s.Variables, struct {
		Key   string
		Value core.Resolved
	}{})
	e = s.CommandEnv()
	if !e.Clear || len(e.Variables) != 2 {
		t.Error("invalid conversion")
	}
}

func TestEnabled(t *testing.T) {
	a := core.Application{}
	if !a.Enabled() {
		t.Error("should be enabled")
	}
	a.Flags = []string{"alojfa"}
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Flags = []string{"pinned"}
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Flags = []string{""}
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Flags = []string{"disabled"}
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Flags = []string{}
	a.Platforms = append(a.Platforms, struct {
		Disable bool
		Value   core.Resolved
		Target  string
	}{})
	if !a.Enabled() {
		t.Error("should be enabled")
	}
	a.Platforms[0].Value = "x"
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Platforms = append(a.Platforms, struct {
		Disable bool
		Value   core.Resolved
		Target  string
	}{false, "a", "a"})
	if !a.Enabled() {
		t.Error("should be enabled")
	}
	a.Platforms[1] = struct {
		Disable bool
		Value   core.Resolved
		Target  string
	}{true, "a", "a"}
	if a.Enabled() {
		t.Error("should be disabled")
	}
}

func TestIsPin(t *testing.T) {
	f := core.FlagSet{}
	if f.Pin() {
		t.Errorf("%s is not pin", f)
	}
	f = core.FlagSet{"pinned"}
	if !f.Pin() {
		t.Errorf("%s is pin", f)
	}
	f = core.FlagSet{"oafoeaj"}
	if f.Pin() {
		t.Errorf("%s is not pin", f)
	}
	f = core.FlagSet{"disabled"}
	if f.Pin() {
		t.Errorf("%s is not pin", f)
	}
}

func TestIsSkipped(t *testing.T) {
	f := core.FlagSet{}
	if f.Skipped() {
		t.Errorf("%s is not skipped", f)
	}
	f = core.FlagSet{"pinned"}
	if !f.Skipped() {
		t.Errorf("%s is skipped", f)
	}
	f = core.FlagSet{"disabled"}
	if !f.Skipped() {
		t.Errorf("%s is skipped", f)
	}
	f = core.FlagSet{"xyz", "disabled", "pinned"}
	if !f.Skipped() {
		t.Errorf("%s is skipped", f)
	}
	f = core.FlagSet{"xyz", "disabled"}
	if !f.Skipped() {
		t.Errorf("%s is skipped", f)
	}
	f = core.FlagSet{"odiajfea"}
	if f.Skipped() {
		t.Errorf("%s is not skipped", f)
	}
}

func TestFlagCheck(t *testing.T) {
	f := core.FlagSet{}
	if err := f.Check(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{""}
	if err := f.Check(); err == nil || err.Error() != "invalid flags, flag set not supported: " {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"pinned"}
	if err := f.Check(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"disabled", "xyz"}
	if err := f.Check(); err == nil || err.Error() != "invalid flags, flag set not supported: disabled,xyz" {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"pinned", "xyz"}
	if err := f.Check(); err == nil || err.Error() != "invalid flags, flag set not supported: pinned,xyz" {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"redeploy", "xyz"}
	if err := f.Check(); err == nil || err.Error() != "invalid flags, flag set not supported: redeploy,xyz" {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"disabled", "pinned"}
	if err := f.Check(); err == nil || err.Error() != "invalid flags, flag set not supported: disabled,pinned" {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"disabled", "redeploy"}
	if err := f.Check(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"pinned", "redeploy"}
	if err := f.Check(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	f = core.FlagSet{"pinned", "redeploy", "disabled"}
	if err := f.Check(); err == nil || err.Error() != "invalid flags, flag set not supported: pinned,redeploy,disabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestWebURL(t *testing.T) {
	s := core.WebURL("abc")
	if s.String() != "abc" || !s.CanTemplate() {
		t.Errorf("invalid web url: %s", s)
	}
}

func TestCommandsFromStep(t *testing.T) {
	s := core.Step{}
	c := slices.Collect(s.Steps())
	if len(c) > 0 {
		t.Errorf("invalid command")
	}
	s.Commands = []interface{}{"x", "y", "z"}
	c = slices.Collect(s.Steps())
	if len(c) != 1 {
		t.Errorf("invalid command")
	}
	s.Commands = []interface{}{[]interface{}{"x"}, []interface{}{"y"}, []interface{}{"z"}}
	c = slices.Collect(s.Steps())
	if len(c) != 3 {
		t.Errorf("invalid command")
	}
}
