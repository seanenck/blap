package config_test

import (
	"testing"

	"github.com/seanenck/blap/internal/config"
)

func TestIndexFile(t *testing.T) {
	f := config.Configuration{}
	f.Directory = "xyz"
	if val := f.IndexFile("abc"); val != "xyz/.blap.abc.index" {
		t.Errorf("invalid index: %s", val)
	}
}
