package core_test

import (
	"testing"

	"github.com/seanenck/blap/internal/core"
)

func TestVersion(t *testing.T) {
	v := core.Version("")
	if v.Major() != "" || v.Minor() != "" || v.Patch() != "" || v.Remainder() != "" || v.Full() != "" {
		t.Errorf("invalid version: %v", v)
	}
	v = core.Version("v1")
	if v.Major() != "1" || v.Minor() != "" || v.Patch() != "" || v.Remainder() != "" || v.Full() != "1" {
		t.Errorf("invalid version: %v", v)
	}
	v = core.Version("1.2")
	if v.Major() != "1" || v.Minor() != "2" || v.Patch() != "" || v.Remainder() != "" || v.Full() != "1.2" {
		t.Errorf("invalid version: %v", v)
	}
	v = core.Version("v1.2.3")
	if v.Major() != "1" || v.Minor() != "2" || v.Patch() != "3" || v.Remainder() != "" || v.Full() != "1.2.3" {
		t.Errorf("invalid version: %v", v)
	}
	v = core.Version("v1.2.3.4.5")
	if v.Major() != "1" || v.Minor() != "2" || v.Patch() != "3" || v.Remainder() != "4.5" || v.Full() != "1.2.3.4.5" {
		t.Errorf("invalid version: %v", v)
	}
}
