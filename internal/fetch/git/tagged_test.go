package git_test

import (
	"net/http"
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/git"
	"github.com/seanenck/blap/internal/fetch/retriever"
)

type mock struct {
	payload []byte
}

func (m *mock) Do(*http.Request) (*http.Response, error) {
	return nil, nil
}

func (m *mock) Output(string, ...string) ([]byte, error) {
	return m.payload, nil
}

func TestTagged(t *testing.T) {
	client := &mock{}
	client.payload = []byte("TESD\ta")
	r := &retriever.ResourceFetcher{}
	r.Backend = client
	if _, err := git.Tagged(r, fetch.Context{Name: "afa"}, core.GitMode{Repository: "{{end}}xyz", Tagged: &core.Filtered{Download: "x", Filters: []string{"a"}}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("TEST\ta")
	r.Backend = client
	if _, err := git.Tagged(r, fetch.Context{Name: "aaa"}, core.GitMode{Repository: "xyz", Tagged: &core.Filtered{Download: "y", Filters: []string{"TEST"}}}); err == nil || err.Error() != "no tags found" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("1\tTEST1\n2\tTEST2\n3\tXYZ3\n4\tZZZ4")
	r.Backend = client
	o, err := git.Tagged(r, fetch.Context{Name: "aljfao"}, core.GitMode{Repository: "a/xyz", Tagged: &core.Filtered{Download: "y/abc", Filters: []string{"(XYZ[0-9])"}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "XYZ3" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "y/abc" || o.File != "abc" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
	client = &mock{}
	client.payload = []byte("TEST\t1\n1\trefs/tags/XYZ2\nXYZ\t3\nZZZ\t4")
	r.Backend = client
	o, err = git.Tagged(r, fetch.Context{Name: "aaa"}, core.GitMode{Repository: "a/xyz", Tagged: &core.Filtered{Download: "xx/s/{{ $.Vars.Tag }}/abc/{{ $.Name }}", Filters: []string{`{{ if ne $.Arch "invalidarch" }}XYZ{{end}}`}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "XYZ" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "xx/s/XYZ/abc/aaa" || o.File != "aaa" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}
