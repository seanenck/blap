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

func TestVersion(t *testing.T) {
	v := fetch.Version("")
	if v.Major() != "" || v.Minor() != "" || v.Patch() != "" || v.Remainder() != "" || v.Version() != "" {
		t.Errorf("invalid version: %v", v)
	}
	v = fetch.Version("v1")
	if v.Major() != "1" || v.Minor() != "" || v.Patch() != "" || v.Remainder() != "" || v.Version() != "1" {
		t.Errorf("invalid version: %v", v)
	}
	v = fetch.Version("1.2")
	if v.Major() != "1" || v.Minor() != "2" || v.Patch() != "" || v.Remainder() != "" || v.Version() != "1.2" {
		t.Errorf("invalid version: %v", v)
	}
	v = fetch.Version("v1.2.3")
	if v.Major() != "1" || v.Minor() != "2" || v.Patch() != "3" || v.Remainder() != "" || v.Version() != "1.2.3" {
		t.Errorf("invalid version: %v", v)
	}
	v = fetch.Version("v1.2.3.4.5")
	if v.Major() != "1" || v.Minor() != "2" || v.Patch() != "3" || v.Remainder() != "4.5" || v.Version() != "1.2.3.4.5" {
		t.Errorf("invalid version: %v", v)
	}
}
