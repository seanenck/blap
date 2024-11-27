package static_test

import (
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/static"
)

func TestNew(t *testing.T) {
	ctx := fetch.Context{}
	ctx.Name = "aa1"
	if _, err := static.New(ctx, core.StaticMode{}); err == nil || err.Error() != "upstream URL not set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := static.New(ctx, core.StaticMode{URL: "abc"}); err == nil || err.Error() != "tag required for static mode" {
		t.Errorf("invalid error: %v", err)
	}
	a, err := static.New(ctx, core.StaticMode{URL: "abc", Tag: "xyz"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if a.URL != "abc" || a.Tag != "xyz" || a.File != "xyz-abc" {
		t.Errorf("invalid resource: %v", a)
	}
	a, err = static.New(ctx, core.StaticMode{URL: "abc", Tag: "xyz", File: "xxx.{{ $.Vars.Tag }}"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if a.URL != "abc" || a.Tag != "xyz" || a.File != "xxx.xyz" {
		t.Errorf("invalid resource: %v", a)
	}
}
