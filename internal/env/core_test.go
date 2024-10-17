package env_test

import (
	"testing"

	"github.com/seanenck/blap/internal/env"
	"github.com/seanenck/blap/internal/util"
)

func TestNewValues(t *testing.T) {
	if _, err := env.NewValues("", struct{}{}); err == nil || err.Error() != "name must be set" {
		t.Errorf("invalid error: %v", err)
	}
	obj, _ := env.NewValues("y", struct{}{})
	if obj.OS == "" {
		t.Errorf("invalid OS: %s", obj.OS)
	}
	if obj.Arch == "" {
		t.Errorf("invalid arch: %s", obj.Arch)
	}
	if util.IsNil(obj) {
		t.Errorf("vars should be nil")
	}
	if obj.Name != "y" {
		t.Errorf("invalid name: %s", obj.Name)
	}
	type testType struct {
		x int
	}
	res, _ := env.NewValues("x", testType{1})
	if res.OS == "" {
		t.Errorf("invalid OS: %s", obj.OS)
	}
	if res.Arch == "" {
		t.Errorf("invalid arch: %s", obj.Arch)
	}
	if res.Vars.x != 1 {
		t.Errorf("invalid application: %d", res.Vars.x)
	}
	if res.Name != "x" {
		t.Errorf("invalid name: %s", obj.Name)
	}
}

func TestTemplate(t *testing.T) {
	o, _ := env.NewValues("abc", struct{ Tag string }{"a"})
	v, err := o.Template("{{ $.Vars.Tag }}{{ $.Name }}")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if v != "aabc" {
		t.Errorf("invalid value: %s", v)
	}
}
