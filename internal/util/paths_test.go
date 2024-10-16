package util_test

import (
	"os"
	"testing"

	"github.com/seanenck/blap/internal/util"
)

func TestPathExists(t *testing.T) {
	if !util.PathExists("paths.go") {
		t.Error("path should exist")
	}
	if util.PathExists("avblakjda") {
		t.Error("path should NOT exist")
	}
}

func TestResolveDirectory(t *testing.T) {
	os.Clearenv()
	if dir := util.ResolveDirectory("abc"); dir != "abc" {
		t.Errorf("invalid dir: %s", dir)
	}
	t.Setenv("HOME", "TEST")
	if dir := util.ResolveDirectory("~/abc"); dir != "TEST/abc" {
		t.Errorf("invalid dir: %s", dir)
	}
}
