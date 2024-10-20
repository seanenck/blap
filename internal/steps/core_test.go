package steps_test

import (
	"testing"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/env"
	"github.com/seanenck/blap/internal/steps"
)

func TestValid(t *testing.T) {
	c := steps.Context{}
	if err := c.Valid(); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables = env.Values[steps.Variables]{}
	if err := c.Valid(); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables.Vars.Resource = &asset.Resource{}
	if err := c.Valid(); err == nil || err.Error() != "directory not set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Variables.Vars.Directory = "abc"
	if err := c.Valid(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
