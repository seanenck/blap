package util_test

import (
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

func TestCleanFileName(t *testing.T) {
	n := util.CleanFileName(" a *#*)!.tar-gz1")
	if n != "a.tar-gz1" {
		t.Errorf("invalid clean: %s", n)
	}
}
