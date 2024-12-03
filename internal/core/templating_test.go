package core_test

import (
	"os"
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/util"
)

func TestNewValues(t *testing.T) {
	if _, err := core.NewValues("", struct{}{}); err == nil || err.Error() != "name must be set" {
		t.Errorf("invalid error: %v", err)
	}
	obj, _ := core.NewValues("y", struct{}{})
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
	res, _ := core.NewValues("x", testType{1})
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
	o, _ := core.NewValues("abc", struct{ Tag string }{"a"})
	defer os.Clearenv()
	t.Setenv("TESTDATA", "TESTING")
	v, err := o.Template("{{ $.Vars.Tag }}{{ $.Name }}{{ $.Getenv \"TESTDATA\" }}")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if v != "aabcTESTING" {
		t.Errorf("invalid value: %s", v)
	}
}
