package cli_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/cli"
)

func TestDefaultConfigs(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	cfgs := cli.DefaultConfigs()
	if len(cfgs) != 0 {
		t.Error("should have no configs")
	}
	t.Setenv("HOME", "xxx")
	cfgs = cli.DefaultConfigs()
	if fmt.Sprintf("%v", cfgs) != "[xxx/.config/blap/config.toml]" {
		t.Errorf("invalid configs: %v", cfgs)
	}
	t.Setenv("XDG_CONFIG_HOME", "yyy")
	cfgs = cli.DefaultConfigs()
	if fmt.Sprintf("%v", cfgs) != "[xxx/.config/blap/config.toml yyy/blap/config.toml]" {
		t.Errorf("invalid configs: %v", cfgs)
	}
}
