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
	c.Resource = env.Values[*asset.Resource]{}
	if err := c.Valid(); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	c.Resource.Vars = &asset.Resource{}
	if err := c.Valid(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
