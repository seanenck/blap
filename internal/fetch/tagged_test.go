package fetch_test

import (
	"testing"

	"github.com/seanenck/blap/internal/fetch"
)

func TestTaggedValidate(t *testing.T) {
	r := &fetch.ResourceFetcher{}
	if _, err := r.Tagged(fetch.TaggedMode{}); err == nil || err.Error() != "no upstream for tagged mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := r.Tagged(fetch.TaggedMode{Repository: "xyz"}); err == nil || err.Error() != "application lacks filters" {
		t.Errorf("invalid error: %v", err)
	}
	client := &mockClient{}
	client.payload = []byte("")
	r.Execute = client
	if _, err := r.Tagged(fetch.TaggedMode{Repository: "xyz", Filters: []string{"TEST"}}); err == nil || err.Error() != "no tags matched" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mockClient{}
	client.payload = []byte("a")
	r.Execute = client
	if _, err := r.Tagged(fetch.TaggedMode{Repository: "xyz", Filters: []string{"TEST"}}); err == nil || err.Error() != "matching version line can not be parsed: a" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestTagged(t *testing.T) {
	client := &mockClient{}
	client.payload = []byte("TESD\ta")
	r := &fetch.ResourceFetcher{}
	r.Execute = client
	if _, err := r.Tagged(fetch.TaggedMode{Repository: "xyz", Filters: []string{"TEST"}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	client = &mockClient{}
	client.payload = []byte("TEST\ta")
	r.Execute = client
	if _, err := r.Tagged(fetch.TaggedMode{Repository: "xyz", Filters: []string{"TEST"}}); err == nil || err.Error() != "no tags matched" {
		t.Errorf("invalid error: %v", err)
	}
	client = &mockClient{}
	client.payload = []byte("TEST\t1\nTEST\t2\nXYZ\t3\nZZZ\t4")
	r.Execute = client
	o, err := r.Tagged(fetch.TaggedMode{Repository: "a/xyz", Filters: []string{"TEST"}})
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
	client = &mockClient{}
	client.payload = []byte("TEST\t1\nTEST\t2\nXYZ\t3\nZZZ\t4")
	r.Execute = client
	o, err = r.Tagged(fetch.TaggedMode{Download: "xx/s", Repository: "a/xyz", Filters: []string{"TEST"}})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if o == nil {
		t.Error("invalid asset")
	} else {
		if o.Tag != "3" {
			t.Errorf("invalid tag: %s", o.Tag)
		}
		if o.URL != "xx/s" || o.File != "s" {
			t.Errorf("invalid asset: %s %s", o.URL, o.File)
		}
	}
}
