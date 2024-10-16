package util_test

import (
	"testing"

	"github.com/seanenck/blap/internal/util"
)

func TestIsNil(t *testing.T) {
	if !util.IsNil(nil) {
		t.Error("is nil")
	}
	var i interface{}
	if !util.IsNil(i) {
		t.Error("is nil")
	}
}
