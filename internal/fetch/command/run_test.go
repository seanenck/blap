package command_test

import (
	"net/http"
	"testing"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/command"
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

func TestRun(t *testing.T) {
	client := &mock{}
	r := &retriever.ResourceFetcher{}
	r.Backend = client
	client.payload = []byte("1.2.3")
	if _, err := command.Run(r, fetch.Context{Name: "afa"}, core.RunMode{Executable: "escho", Arguments: []string{"1.2.3"}, Fetch: &core.Filtered{Sort: "", Download: "ajfaeaijo", Filters: []string{"(.*?)"}}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("v0.1.2\n2.3.0")
	r.Backend = client
	if _, err := command.Run(r, fetch.Context{Name: "afa"}, core.RunMode{Executable: "echeo", Arguments: nil, Fetch: &core.Filtered{Sort: "", Download: "ajfaeaijo", Filters: []string{"(TEST?)"}}}); err == nil || err.Error() != "no tags found" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mock{}
	client.payload = []byte("abc-0.1.2.txt\nabc-2.3.0.txt\n\nabc-aaa-1.2.3\nabc-1.1.2.txt")
	r.Backend = client
	o, err := command.Run(r, fetch.Context{Name: "aljfao"}, core.RunMode{Executable: "echo", Arguments: []string{"aojoeaj"}, Fetch: &core.Filtered{Sort: "semver", Download: "oijoefa/x", Filters: []string{"abc-([0-9.]*?).txt"}}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "v2.3.0" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "oijoefa/x" || o.File != "x" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}