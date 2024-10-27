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

func TestTaggedValidate(t *testing.T) {
	r := &retriever.ResourceFetcher{}
	if _, err := git.Tagged(r, fetch.Context{Name: "abc"}, core.GitMode{}); err == nil || err.Error() != "tagged definition is nil" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := git.Tagged(r, fetch.Context{Name: "xyz"}, core.GitMode{Tagged: &core.GitTaggedMode{}}); err == nil || err.Error() != "no upstream for tagged mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := git.Tagged(r, fetch.Context{Name: "xxx"}, core.GitMode{Repository: "xyz", Tagged: &core.GitTaggedMode{}}); err == nil || err.Error() != "application lacks filters" {
		t.Errorf("invalid error: %v", err)
	}
	client := &mock{}
	client.payload = []byte("")
	r.Backend = client
	if _, err := git.Tagged(r, fetch.Context{Name: "j1o2i"}, core.GitMode{Repository: "xyz", Tagged: &core.GitTaggedMode{Filters: []string{"TEST"}}}); err == nil || err.Error() != "no tags matched" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("a")
	r.Backend = client
	if _, err := git.Tagged(r, fetch.Context{Name: "xxx"}, core.GitMode{Repository: "xyz", Tagged: &core.GitTaggedMode{Filters: []string{"TEST"}}}); err == nil || err.Error() != "matching version line can not be parsed: a" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestTagged(t *testing.T) {
	client := &mock{}
	client.payload = []byte("TESD\ta")
	r := &retriever.ResourceFetcher{}
	r.Backend = client
	if _, err := git.Tagged(r, fetch.Context{Name: "afa"}, core.GitMode{Repository: "xyz", Tagged: &core.GitTaggedMode{Filters: []string{"TEST"}}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("TEST\ta")
	r.Backend = client
	if _, err := git.Tagged(r, fetch.Context{Name: "aaa"}, core.GitMode{Repository: "xyz", Tagged: &core.GitTaggedMode{Filters: []string{"TEST"}}}); err == nil || err.Error() != "no tags matched" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("TEST\t1\nTEST\t2\nXYZ\t3\nZZZ\t4")
	r.Backend = client
	o, err := git.Tagged(r, fetch.Context{Name: "aljfao"}, core.GitMode{Repository: "a/xyz", Tagged: &core.GitTaggedMode{Filters: []string{"TEST"}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "3" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "a/xyz" || o.File != "xyz" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
	client = &mock{}
	client.payload = []byte("TEST\t1\nTEST\t2\nXYZ\t3\nZZZ\t4")
	r.Backend = client
	o, err = git.Tagged(r, fetch.Context{Name: "aaa"}, core.GitMode{Repository: "a/xyz", Tagged: &core.GitTaggedMode{Download: "xx/s/{{ $.Vars.Tag }}/abc/{{ $.Name }}", Filters: []string{`{{ if ne $.Arch "invalidarch" }}TEST{{end}}`}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "3" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "xx/s/3/abc/aaa" || o.File != "aaa" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
	if _, err = git.Tagged(r, fetch.Context{}, core.GitMode{Repository: "a/xyz", Tagged: &core.GitTaggedMode{Filters: []string{"TEST"}}}); err == nil || err.Error() != "context missing name" {
		t.Errorf("invalid error: %v", err)
	}
}
