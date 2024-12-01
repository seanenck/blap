package processing_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/processing"
)

func makeTestFile(src string) {
	b, _ := os.ReadFile(filepath.Join("examples", src))
	to := filepath.Join("testdata", "test.toml")
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	os.WriteFile(to, b, 0o644)
}

func TestLoadError(t *testing.T) {
	makeTestFile("disabled.more.toml")
	example := filepath.Join("examples", "config.toml")
	if _, err := processing.Load(example, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	makeTestFile("added.more.toml")
	if _, err := processing.Load(example, cli.Settings{}); err == nil || !strings.Contains(err.Error(), "is overwritten by config:") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestLoad(t *testing.T) {
	makeTestFile("disabled.more.toml")
	example := filepath.Join("examples", "config.toml")
	c, err := processing.Load(example, cli.Settings{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 8 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
	if len(c.Pinned) != 4 {
		t.Errorf("invalid pins: %d", len(c.Pinned))
	}
}

func TestLoadFilter(t *testing.T) {
	makeTestFile("disabled.more.toml")
	example := filepath.Join("examples", "config.toml")
	s := cli.Settings{}
	s.CompileApplicationFilter("l", true)
	c, err := processing.Load(example, s)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 5 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
	s.CompileApplicationFilter("l", false)
	c, err = processing.Load(example, s)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 3 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
}

func TestLoadInclude(t *testing.T) {
	makeTestFile("disabled.more.toml")
	example := filepath.Join("examples", "config.toml")
	s := cli.Settings{}
	re, _ := regexp.Compile("other")
	s.Include = re
	c, err := processing.Load(example, s)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(c.Applications) != 3 {
		t.Errorf("invalid apps: %d", len(c.Applications))
	}
}
