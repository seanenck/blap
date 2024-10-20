package types_test

import (
	"fmt"
	"os"
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

func TestResolve(t *testing.T) {
	os.Clearenv()
	v := types.Resolved("")
	if v.String() != "" {
		t.Errorf("invalid result: %v", v)
	}
	t.Setenv("HOME", "abc")
	v = types.Resolved("~/")
	if v.String() != "abc" {
		t.Errorf("invalid result: %v", v)
	}
	t.Setenv("XXX", "123")
	v = types.Resolved("~/$XXX")
	if v.String() != "abc/123" {
		t.Errorf("invalid result: %v", v)
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

func TestVarSetUnset(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	val := types.Variables{}
	val.Set()
	val.Unset()
	val.Vars = make(map[string]string)
	val.Set()
	val.Unset()
	os.Setenv("HOME", "1")
	val.Vars["HOME"] = "2"
	val.Vars["THIS_IS_A_TEST"] = "3"
	val.Set()
	if os.Getenv("HOME") != "2" || os.Getenv("THIS_IS_A_TEST") != "3" {
		t.Errorf("invalid env")
	}
	val.Unset()
	if os.Getenv("HOME") != "1" || os.Getenv("THIS_IS_A_TEST") != "" {
		t.Errorf("invalid env")
	}
}
