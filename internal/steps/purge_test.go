package steps_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/seanenck/blap/internal/steps"
)

func TestPurge(t *testing.T) {
	did := false
	v := ""
	isDryRun := false
	fxn := func(value string) bool {
		did = true
		v = value
		return !isDryRun
	}
	if err := steps.Purge("", nil, nil, fxn); err == nil || err.Error() != "directory must be set" || did {
		t.Errorf("invalid error: %v|purge", err)
	}
	did = false
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	if err := steps.Purge("testdata", nil, nil, fxn); err != nil || did {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 0 {
		t.Errorf("invalid dirs: %v", d)
	}
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "xyz"), 0o755)
	did = false
	isDryRun = true
	if err := steps.Purge("testdata", []string{"abc"}, nil, fxn); err != nil || !did || v != "xyz" {
		t.Errorf("invalid error: %v|purge (%s)", err, v)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 2 {
		t.Errorf("invalid dirs: %v", d)
	}
	did = false
	v = ""
	p, _ := regexp.Compile("xyz")
	isDryRun = false
	if err := steps.Purge("testdata", []string{"abc"}, []*regexp.Regexp{p}, fxn); err != nil || did {
		t.Errorf("invalid error: %v|purge", err)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 2 {
		t.Errorf("invalid dirs: %v", d)
	}
	did = false
	if err := steps.Purge("testdata", []string{"abc"}, nil, fxn); err != nil || !did || v != "xyz" {
		t.Errorf("invalid error: %v|purge (%s)", err, v)
	}
	if d, _ := os.ReadDir("testdata"); len(d) != 1 {
		t.Errorf("invalid dirs: %v", d)
	}
}
