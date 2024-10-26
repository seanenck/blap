package purge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/purge"
)

func TestDo(t *testing.T) {
	did := false
	fxn := func() {
		did = true
	}
	if err := purge.Do("", nil, nil, cli.Settings{}, fxn); err == nil || err.Error() != "directory must be set" || did {
		t.Errorf("invalid error: %v|purge", err)
	}
	did = false
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	if err := purge.Do("testdata", nil, nil, cli.Settings{}, fxn); err != nil || did {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 0 {
		t.Errorf("invalid dirs: %v", d)
	}
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "xyz"), 0o755)
	did = false
	if err := purge.Do("testdata", []string{"abc"}, nil, cli.Settings{DryRun: true}, fxn); err != nil || !did {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 2 {
		t.Errorf("invalid dirs: %v", d)
	}
	did = false
	if err := purge.Do("testdata", []string{"abc"}, []string{"xyz"}, cli.Settings{}, fxn); err != nil || did {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 2 {
		t.Errorf("invalid dirs: %v", d)
	}
	did = false
	if err := purge.Do("testdata", []string{"abc"}, nil, cli.Settings{}, fxn); err != nil || !did {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 1 {
		t.Errorf("invalid dirs: %v", d)
	}
}
