// Package steps are common process step definitions
package steps

import (
	"errors"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/env"
)

type (
	// Context are step settings/context
	Context struct {
		Settings cli.Settings
		Resource env.Values[*asset.Resource]
	}
)

// Valid will check validity of the context
func (c Context) Valid() error {
	if c.Resource.Vars == nil {
		return errors.New("no resource set")
	}
	return nil
}
