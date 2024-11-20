package core_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/core"
)

func TestSourceItems(t *testing.T) {
	s := core.Application{}
	cnt := 0
	for range s.Items() {
		cnt++
	}
	if cnt != 3 {
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
	a.Commands.ClearEnv = true
	a.Commands.Variables = append(a.Commands.Variables, struct {
		Key   string
		Value core.Resolved
	}{})
	e := a.CommandEnv()
	if !e.Clear || len(e.Variables) != 1 {
		t.Error("invalid conversion")
	}
	s := core.Application{}
	s.Commands.ClearEnv = true
	s.Commands.Variables = append(s.Commands.Variables, struct {
		Key   string
		Value core.Resolved
	}{})
	s.Commands.Variables = append(s.Commands.Variables, struct {
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
	a.Disable = true
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Disable = false
	a.Platforms = append(a.Platforms, struct {
		Value  core.Resolved
		Target string
	}{})
	if !a.Enabled() {
		t.Error("should be enabled")
	}
	a.Platforms[0].Value = "x"
	if a.Enabled() {
		t.Error("should be disabled")
	}
	a.Platforms = append(a.Platforms, struct {
		Value  core.Resolved
		Target string
	}{"a", "a"})
	if !a.Enabled() {
		t.Error("should be enabled")
	}
}
