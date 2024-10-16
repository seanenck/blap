package cli_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/cli"
)

func TestLogging(t *testing.T) {
	c := cli.Settings{}
	c.LogCore("abc")
	var buf bytes.Buffer
	c.Writer = &buf
	c.LogCore("xyz")
	s := buf.String()
	if s != "" {
		t.Errorf("invalid buffer result: %s", s)
	}
	buf = bytes.Buffer{}
	c.Writer = &buf
	c.Verbosity = 1
	c.LogCore("a11")
	c.LogDebug("xyz")
	c.LogInfo("zbc")
	s = buf.String()
	if s != "a11" {
		t.Errorf("invalid buffer result: %s", s)
	}
	buf = bytes.Buffer{}
	c.Writer = &buf
	c.Verbosity = 2
	c.LogCore("a11")
	c.LogDebug("xyz")
	c.LogInfo("zbc")
	s = buf.String()
	if s != "a11zbc" {
		t.Errorf("invalid buffer result: %s", s)
	}
	buf = bytes.Buffer{}
	c.Writer = &buf
	c.Verbosity = 100
	c.LogCore("a11")
	c.LogDebug("xyz")
	c.LogInfo("zbc")
	s = buf.String()
	if s != "a11xyzzbc" {
		t.Errorf("invalid buffer result: %s", s)
	}
}

func TestCompileFilter(t *testing.T) {
	c := cli.Settings{}
	if err := c.CompileApplicationFilter("", false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := c.CompileApplicationFilter("x", false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := c.CompileApplicationFilter("*", false); err == nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestFiltered(t *testing.T) {
	c := cli.Settings{}
	c.CompileApplicationFilter("", false)
	if c.FilterApplications() {
		t.Error("is NOT filtered")
	}
	c.CompileApplicationFilter("a", false)
	if !c.FilterApplications() {
		t.Error("is filtered")
	}
}

func TestAllowed(t *testing.T) {
	c := cli.Settings{}
	c.CompileApplicationFilter("", false)
	if !c.AllowApplication("ajoa") {
		t.Error("should be allowed")
	}
	c.CompileApplicationFilter("", true)
	if !c.AllowApplication("ajoa") {
		t.Error("should be allowed")
	}
	c.CompileApplicationFilter("a", false)
	if !c.AllowApplication("ajoa") {
		t.Error("should be allowed")
	}
	c.CompileApplicationFilter("a", true)
	if c.AllowApplication("ajoa") {
		t.Error("should NOT be allowed")
	}
}

func TestResolveDir(t *testing.T) {
	os.Clearenv()
	c := cli.Settings{Resolves: make(map[string]string)}
	if dir := c.Resolve("abc"); dir != "abc" {
		t.Errorf("invalid dir: %s", dir)
	}
	os.Unsetenv("HOME")
	if dir := c.Resolve("~/abc"); dir != "~/abc" {
		t.Errorf("invalid dir: %s", dir)
	}
	t.Setenv("HOME", "TEST")
	if dir := c.Resolve("~/abc"); dir != "~/abc" {
		t.Errorf("invalid dir: %s", dir)
	}
	t.Setenv("HOME", "TEST")
	if dir := c.Resolve("~/ix"); dir != "TEST/ix" {
		t.Errorf("invalid dir: %s", dir)
	}
}
