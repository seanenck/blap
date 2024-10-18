package types_test

import (
	"os"
	"strings"
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

func TestGitHubToken(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	r, err := types.ParseToken(types.GitHubSettings{})
	if r != "" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = types.ParseToken(types.GitHubSettings{Token: "abc"})
	if r != "abc" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	r, err = types.ParseToken(types.GitHubSettings{Token: "core_test.go"})
	if r == "" || !strings.Contains(r, "package types_test") || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	t.Setenv("GITHUB_TOKEN", "xyz")
	r, err = types.ParseToken(types.GitHubSettings{Token: "abc"})
	if r != "xyz" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
	t.Setenv("BLAP_GITHUB_TOKEN", "123")
	r, err = types.ParseToken(types.GitHubSettings{Token: "abc"})
	if r != "123" || err != nil {
		t.Errorf("invalid result: %s %v", r, err)
	}
}
