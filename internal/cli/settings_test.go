package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/logging"
)

func TestLogging(t *testing.T) {
	c := cli.Settings{}
	c.LogCore(logging.BuildCategory, "abc")
	var buf bytes.Buffer
	c.Writer = &buf
	c.LogCore(logging.FetchCategory, "xyz")
	s := buf.String()
	if s != "" {
		t.Errorf("invalid buffer result: %s", s)
	}
	buf = bytes.Buffer{}
	c.Writer = &buf
	c.Verbosity = 1
	c.LogCore(logging.BuildCategory, "a11")
	c.LogDebug(logging.BuildCategory, "xyz")
	s = buf.String()
	if s != "[build] a11" {
		t.Errorf("invalid buffer result: %s", s)
	}
	buf = bytes.Buffer{}
	c.Writer = &buf
	c.Verbosity = 2
	c.LogCore(logging.FetchCategory, "a11")
	c.LogDebug(logging.FetchCategory, "xyz")
	s = buf.String()
	if s != "[fetch] a11" {
		t.Errorf("invalid buffer result: %s", s)
	}
	buf = bytes.Buffer{}
	c.Writer = &buf
	c.Verbosity = 100
	c.LogCore(logging.IndexCategory, "a11")
	c.LogDebug(logging.BuildCategory, "xyz")
	s = buf.String()
	if s != "[index] a11[build] xyz" {
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

func TestParseToken(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	s := cli.Settings{}
	r, err := s.ParseToken(core.GitHubSettings{})
	if r != "" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = s.ParseToken(core.GitHubSettings{Token: "abc"})
	if r != "abc" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	os.Mkdir("testdata", 0o755)
	test := filepath.Join("testdata", "test.sh")
	os.WriteFile(test, []byte("#!/bin/sh\necho 123 $1"), 0o755)
	r, err = s.ParseToken(core.GitHubSettings{Command: []core.Resolved{core.Resolved(test)}})
	if r != "123" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = s.ParseToken(core.GitHubSettings{Token: "111"})
	if r != "111" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = s.ParseToken(core.GitHubSettings{Token: "111", Command: []core.Resolved{"x"}})
	if r != "111" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	if _, err := s.ParseToken(core.GitHubSettings{Command: []core.Resolved{"x"}}); err == nil || !strings.Contains(err.Error(), "executable file not") {
		t.Errorf("invalid result: %v", err)
	}
	t.Setenv("GITHUB_TOKEN", "xyz")
	r, err = s.ParseToken(core.GitHubSettings{Token: "abc"})
	if r != "xyz" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	t.Setenv("BLAP_GITHUB_TOKEN", "123")
	r, err = s.ParseToken(core.GitHubSettings{Token: "abc"})
	if r != "123" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
}
