package processing_test

import (
	"testing"

	"github.com/seanenck/blap/internal/processing"
)

func TestIndexFile(t *testing.T) {
	f := processing.Configuration{}
	f.Directory = "xyz"
	if val := f.IndexFile("abc"); val != "xyz/.blap.abc.index" {
		t.Errorf("invalid index: %s", val)
	}
}