package fetch_test

import (
	"testing"

	"github.com/seanenck/blap/internal/fetch"
)

func TestTemplating(t *testing.T) {
	ctx := fetch.Context{}
	if _, err := ctx.Templating("", nil); err == nil || err.Error() != "context missing name" {
		t.Errorf("invalid error: %v", err)
	}
	ctx.Name = "xyz"
	res, err := ctx.Templating("{{ $.Name }}{{ $.Vars.Tag }}", &fetch.Template{Tag: "xz"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if res != "xyzxz" {
		t.Errorf("invalid result: %s", res)
	}
	res, err = ctx.Templating("{{ $.Name }}{{ $.Vars.Tag }}", &fetch.Template{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if res != "xyz" {
		t.Errorf("invalid result: %s", res)
	}
	res, err = ctx.Templating("{{ $.Name }}", nil)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if res != "xyz" {
		t.Errorf("invalid result: %s", res)
	}
}

func TestCompileRegex(t *testing.T) {
	ctx := fetch.Context{}
	ctx.Name = "xyz"
	r, err := ctx.CompileRegexp("{{ $.Vars.Tag }}", &fetch.Template{Tag: "abc"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !r.MatchString("abc") {
		t.Error("regex should match")
	}
}
