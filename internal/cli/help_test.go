package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
)

func TestUsage(t *testing.T) {
	if err := cli.Usage(nil); err == nil || err.Error() != "nil writer" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	t.Setenv("HOME", "xxxxx")
	t.Setenv("XDG_CONFIG_HOME", "yyyyy")
	var buf bytes.Buffer
	if err := cli.Usage(&buf); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "xxxxx") || !strings.Contains(s, "yyyyy") {
		t.Errorf("invalid help: %s", s)
	}
}
