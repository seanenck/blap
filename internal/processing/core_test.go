package processing_test

import (
	"testing"

	"github.com/seanenck/blap/internal/processing"
)

func TestIndexFile(t *testing.T) {
	f := processing.Configuration{}
	if val := f.IndexFile("abc"); val != ".blap.abc.index" {
		t.Errorf("invalid index: %s", val)
	}
}

func TestNewFile(t *testing.T) {
	f := processing.Configuration{}
	if val := f.NewFile("abc"); val != "abc" {
		t.Errorf("invalid index: %s", val)
	}
}
