package deploy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/env"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/steps/deploy"
)

func TestDo(t *testing.T) {
	step := steps.Context{}
	step.Settings = cli.Settings{}
	if err := deploy.Do("", nil, step); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	obj := types.Artifact{}
	objs := types.Artifact{}
	if err := deploy.Do("", []types.Artifact{objs, obj}, step); err == nil || err.Error() != "source directory required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := deploy.Do("testdata", []types.Artifact{objs, obj}, step); err == nil || err.Error() != "no resource set" {
		t.Errorf("invalid error: %v", err)
	}
	e, _ := env.NewValues("a", &asset.Resource{})
	step.Resource = e
	if err := deploy.Do("testdata", []types.Artifact{objs, obj}, step); err == nil || err.Error() != "missing deploy destination" {
		t.Errorf("invalid error: %v", err)
	}
	dest := filepath.Join("testdata", "dir")
	obj.Destination = strings.ReplaceAll(dest, "a", "{{ $.Name }}")
	objs.Destination = dest
	os.Mkdir(dest, 0o755)
	if err := deploy.Do("testdata", []types.Artifact{objs, obj}, step); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	obj.Files = append(obj.Files, "")
	if err := deploy.Do("testdata", []types.Artifact{objs, obj}, step); err == nil || err.Error() != "empty file not allowed" {
		t.Errorf("invalid error: %v", err)
	}
	obj.Files = []string{}
	obj.Files = append(obj.Files, "{{ $.Name }}")
	if err := deploy.Do("testdata", []types.Artifact{objs, obj}, step); err == nil || err.Error() != "unable to find source file: testdata/a" {
		t.Errorf("invalid error: %v", err)
	}
	testFile := filepath.Join(dest, "a")
	os.Remove(testFile)
	os.WriteFile(filepath.Join("testdata", "a"), []byte{}, 0o644)
	os.WriteFile(filepath.Join("testdata", "b"), []byte{}, 0o644)
	os.WriteFile(testFile, []byte{}, 0o644)
	objs.Files = append(obj.Files, "b")
	if err := deploy.Do("testdata", []types.Artifact{objs, obj}, step); err != nil {
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
