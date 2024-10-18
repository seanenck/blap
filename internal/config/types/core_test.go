package types_test

import (
	"fmt"
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
	token := types.GitHubSettings{}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if token.Value() != "" {
		t.Errorf("invalid value: %s", token.Value())
	}
	token = types.GitHubSettings{Token: "xyz"}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if token.Value() != "xyz" {
		t.Errorf("invalid value: %s", token.Value())
	}
}
