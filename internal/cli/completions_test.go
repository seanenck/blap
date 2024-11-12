package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
)

func TestGenerationCompletions(t *testing.T) {
	if err := cli.GenerateCompletions(nil); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	var buf bytes.Buffer
	t.Setenv("SHELL", "x/a")
	if err := cli.GenerateCompletions(&buf); err == nil || err.Error() != "unable to generate completions for shell" {
		t.Errorf("invalid error: %v", err)
	}
	buf = bytes.Buffer{}
	t.Setenv("SHELL", "bash")
	if err := cli.GenerateCompletions(&buf); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	b := buf.String()
	if !strings.Contains(b, "local ") {
		t.Errorf("invalid buffer: %s", b)
	}
	t.Setenv("SHELL", "zsh")
	if err := cli.GenerateCompletions(&buf); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	b = buf.String()
	if !strings.Contains(b, "main") {
		t.Errorf("invalid buffer: %s", b)
	}
}
