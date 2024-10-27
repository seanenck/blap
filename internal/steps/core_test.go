package steps_test

import (
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/steps"
)

func TestValid(t *testing.T) {
	c := steps.Context{}
	if err := c.Valid(); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables = core.Values[steps.Variables]{}
	if err := c.Valid(); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables.Vars.Resource = &core.Resource{}
	if err := c.Valid(); err == nil || err.Error() != "directory not set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables.Vars.Directories.Root = "abc"
	if err := c.Valid(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestClone(t *testing.T) {
	v := steps.Variables{}
	v.Clone()
	v.Directories.Root = "xyz"
	v.Resource = &core.Resource{File: "y"}
	v.Directories.Working = "work"
	n := v.Clone()
	if n.Directories.Root != "xyz" || n.Directories.Working != "work" || n.File != "y" {
		t.Error("invalid clone")
	}
}
