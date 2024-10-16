package deploy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/deploy"
)

func TestDo(t *testing.T) {
	if err := deploy.Do("", nil, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	obj := deploy.Artifact{}
	objs := deploy.Artifact{}
	if err := deploy.Do("", []deploy.Artifact{objs, obj}, cli.Settings{}); err == nil || err.Error() != "source directory required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := deploy.Do("testdata", []deploy.Artifact{objs, obj}, cli.Settings{}); err == nil || err.Error() != "missing deploy destination" {
		t.Errorf("invalid error: %v", err)
	}
	dest := filepath.Join("testdata", "dir")
	obj.Destination = dest
	objs.Destination = dest
	os.Mkdir(dest, 0o755)
	if err := deploy.Do("testdata", []deploy.Artifact{objs, obj}, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	obj.Files = append(obj.Files, "")
	if err := deploy.Do("testdata", []deploy.Artifact{objs, obj}, cli.Settings{}); err == nil || err.Error() != "empty file not allowed" {
		t.Errorf("invalid error: %v", err)
	}
	obj.Files = []string{}
	obj.Files = append(obj.Files, "a")
	if err := deploy.Do("testdata", []deploy.Artifact{objs, obj}, cli.Settings{}); err == nil || err.Error() != "unable to find source file: testdata/a" {
		t.Errorf("invalid error: %v", err)
	}
	testFile := filepath.Join(dest, "a")
	os.Remove(testFile)
	os.WriteFile(filepath.Join("testdata", "a"), []byte{}, 0o644)
	os.WriteFile(filepath.Join("testdata", "b"), []byte{}, 0o644)
	os.WriteFile(testFile, []byte{}, 0o644)
	objs.Files = append(obj.Files, "b")
	if err := deploy.Do("testdata", []deploy.Artifact{objs, obj}, cli.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	for k, d := range map[string]int{
		"testdata": 3,
		dest:       2,
	} {
		if dirs, _ := os.ReadDir(k); len(dirs) != d {
			t.Errorf("invalid result: %s %d", k, len(dirs))
		}
	}
}
