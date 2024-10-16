package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config"
)

func makeTestFile(src string) {
	b, _ := os.ReadFile(filepath.Join("examples", src))
	to := filepath.Join("testdata", "test.yaml")
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	os.WriteFile(to, b, 0o644)
}

func TestLoadError(t *testing.T) {
	makeTestFile("disabled.more.yaml")
	example := filepath.Join("examples", "config.yaml")
	if _, err := config.Load(example, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	makeTestFile("added.more.yaml")
	if _, err := config.Load(example, cli.Settings{}); err == nil || !strings.Contains(err.Error(), "is overwritten by config:") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestLoad(t *testing.T) {
	makeTestFile("disabled.more.yaml")
	example := filepath.Join("examples", "config.yaml")
	c, err := config.Load(example, cli.Settings{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 5 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
}

func TestLoadFilter(t *testing.T) {
	makeTestFile("disabled.more.yaml")
	example := filepath.Join("examples", "config.yaml")
	s := cli.Settings{}
	s.CompileApplicationFilter("l", true)
	c, err := config.Load(example, s)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 3 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
	s.CompileApplicationFilter("l", false)
	c, err = config.Load(example, s)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 2 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
}
