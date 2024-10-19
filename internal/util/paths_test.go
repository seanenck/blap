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
