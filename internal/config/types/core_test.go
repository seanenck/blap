package types_test

import (
	"testing"

	"github.com/seanenck/blap/internal/config/types"
)

func TestSourceItems(t *testing.T) {
	s := types.Source{}
	cnt := 0
	for range s.Items() {
		cnt++
	}
	if cnt != 2 {
		t.Errorf("invalid reflection count %d", cnt)
	}
}
