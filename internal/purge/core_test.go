package purge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/purge"
)

func TestDo(t *testing.T) {
	if ok, err := purge.Do("", nil, cli.Settings{}); err == nil || err.Error() != "directory must be set" || ok {
		t.Errorf("invalid error: %v|purge", err)
	}
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	if ok, err := purge.Do("testdata", nil, cli.Settings{}); err != nil || ok {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 0 {
		t.Errorf("invalid dirs: %v", d)
	}
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "xyz"), 0o755)
	if ok, err := purge.Do("testdata", []string{"abc"}, cli.Settings{DryRun: true}); err != nil || !ok {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 2 {
		t.Errorf("invalid dirs: %v", d)
	}
	if ok, err := purge.Do("testdata", []string{"abc"}, cli.Settings{}); err != nil || !ok {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 1 {
		t.Errorf("invalid dirs: %v", d)
	}
}
