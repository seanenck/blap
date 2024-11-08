package core_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/blap/internal/core"
)

func TestSourceItems(t *testing.T) {
	s := core.Source{}
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
	v := core.Resolved("")
	if v.String() != "" {
		t.Errorf("invalid result: %v", v)
	}
	t.Setenv("HOME", "abc")
	v = core.Resolved("~/")
	if v.String() != "abc" {
		t.Errorf("invalid result: %v", v)
	}
	t.Setenv("XXX", "123")
	v = core.Resolved("~/$XXX")
	if v.String() != "abc/123" {
		t.Errorf("invalid result: %v", v)
	}
}

func TestGitHubToken(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	token := core.GitHubSettings{}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if token.Value() != nil {
		t.Errorf("invalid value: %s", token.Value())
	}
	token = core.GitHubSettings{Token: "xyz"}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if fmt.Sprintf("%v", token.Value()) != "[xyz]" {
		t.Errorf("invalid value: %s", token.Value())
	}
	t.Setenv("HOME", "zzz")
	token = core.GitHubSettings{Token: []interface{}{"$HOME/xyz", "$HOME/111"}}
	if fmt.Sprintf("%v", token.Env()) != "[BLAP_GITHUB_TOKEN GITHUB_TOKEN]" {
		t.Errorf("invalid token: %v", token.Env())
	}
	if fmt.Sprintf("%v", token.Value()) != "[zzz/xyz zzz/111]" {
		t.Errorf("invalid value: %s", token.Value())
	}
}

func TestVarSetUnset(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	val := core.Variables{}
	val.Set()
	val.Unset()
	val.Vars = make(map[string]core.Resolved)
	val.Set()
	val.Unset()
	os.Setenv("HOME", "1")
	os.Setenv("A_TEST", "0")
	val.Vars["A_TEST"] = "~/2"
	val.Vars["THIS_IS_A_TEST"] = "3"
	val.Set()
	if os.Getenv("A_TEST") != "1/2" || os.Getenv("THIS_IS_A_TEST") != "3" {
		t.Errorf("invalid env")
	}
	val.Unset()
	if os.Getenv("A_TEST") != "0" || os.Getenv("THIS_IS_A_TEST") != "" {
		t.Errorf("invalid env")
	}
}
