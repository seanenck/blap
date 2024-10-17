package github_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/seanenck/blap/internal/fetch/github"
)

type mock struct {
	req     *http.Request
	payload []byte
}

func (m *mock) Output(string, ...string) ([]byte, error) {
	return nil, nil
}

func (m *mock) Do(r *http.Request) (*http.Response, error) {
	m.req = r
	length := len(m.payload)
	if length > 0 {
		resp := &http.Response{}
		resp.Body = io.NopCloser(bytes.NewBuffer(m.payload))
		resp.ContentLength = int64(length)
		resp.StatusCode = http.StatusOK
		return resp, nil
	}
	return nil, nil
}

func TestWrapperError(t *testing.T) {
	w := github.WrapperError{}
	if w.Error() != "code: 0" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.URL = "a"
	if w.Error() != "code: 0\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Status = "x"
	if w.Error() != "code: 0\nstatus: x\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Body = []byte("[")
	if w.Error() != "code: 0\nstatus: x\nunmarshal: unexpected end of JSON input\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Body = []byte("{}")
	if w.Error() != "code: 0\nstatus: x\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
	w.Body = []byte(`{"message": "mess", "documentation_url": "xxx"}`)
	if w.Error() != "code: 0\ndoc: xxx\nmessage: mess\nstatus: x\nurl: a" {
		t.Errorf("invalid error: %s", w.Error())
	}
}
