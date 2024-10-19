package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
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

func TestParseToken(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	s := cli.Settings{}
	r, err := s.ParseToken(types.GitHubSettings{})
	if r != "" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = s.ParseToken(types.GitHubSettings{Token: "abc"})
	if r != "abc" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = s.ParseToken(types.GitHubSettings{Token: "settings_test.go"})
	if r == "" || !strings.Contains(r, "package cli_test") || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	t.Setenv("GITHUB_TOKEN", "xyz")
	r, err = s.ParseToken(types.GitHubSettings{Token: "abc"})
	if r != "xyz" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	t.Setenv("BLAP_GITHUB_TOKEN", "123")
	r, err = s.ParseToken(types.GitHubSettings{Token: "abc"})
	if r != "123" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
}
