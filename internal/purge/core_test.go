package purge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/purge"
)

func TestDo(t *testing.T) {
	if err := purge.Do("", nil, cli.Settings{}); err == nil || err.Error() != "directory must be set" {
		t.Errorf("invalid error: %v", err)
	}
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	if err := purge.Do("testdata", nil, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 0 {
		t.Errorf("invalid dirs: %v", d)
	}
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "xyz"), 0o755)
	if err := purge.Do("testdata", []string{"abc"}, cli.Settings{DryRun: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 2 {
		t.Errorf("invalid dirs: %v", d)
	}
	if err := purge.Do("testdata", []string{"abc"}, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 1 {
		t.Errorf("invalid dirs: %v", d)
	}
}
