package asset_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/asset"
)

type mockExtract struct {
	payload []byte
	err     error
	ran     []string
}

func (m *mockExtract) LogDebug(string, ...any) {
}

func (m *mockExtract) Run(c string, a ...string) error {
	m.ran = []string{c}
	m.ran = append(m.ran, a...)
	return m.err
}

func (m *mockExtract) RunIn(string, string, ...string) error {
	return nil
}

func (m *mockExtract) Output(c string, a ...string) ([]byte, error) {
	return m.payload, m.Run(c, a...)
}

func TestSetAppDataErrors(t *testing.T) {
	r := &asset.Resource{}
	if err := r.SetAppData("", "", asset.Settings{}); err == nil || err.Error() != "name and directory are required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := r.SetAppData("a", "", asset.Settings{}); err == nil || err.Error() != "name and directory are required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := r.SetAppData("", "b", asset.Settings{}); err == nil || err.Error() != "name and directory are required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := r.SetAppData("a", "b", asset.Settings{}); err == nil || err.Error() != "asset not initialized properly" {
		t.Errorf("invalid error: %v", err)
	}
	r.File = "x"
	if err := r.SetAppData("a", "b", asset.Settings{}); err == nil || err.Error() != "asset not initialized properly" {
		t.Errorf("invalid error: %v", err)
	}
	r.Tag = "x"
	if err := r.SetAppData("a", "b", asset.Settings{}); err == nil || err.Error() != "asset not initialized properly" {
		t.Errorf("invalid error: %v", err)
	}
	r.URL = "x"
	if err := r.SetAppData("a", "b", asset.Settings{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestExtractErrors(t *testing.T) {
	r := &asset.Resource{}
	r.File = "file"
	r.Tag = "tag"
	r.URL = "url"
	if err := r.Extract(&mockExtract{}); err == nil || err.Error() != "asset not set for extraction" {
		t.Errorf("invalid error: %v", err)
	}
	r.SetAppData("a", "b", asset.Settings{})
	if err := r.Extract(&mockExtract{}); err == nil || err.Error() != "asset has no extraction command" {
		t.Errorf("invalid error: %v", err)
	}
	r.SetAppData("a", "b", asset.Settings{Command: []string{"xyz"}})
	if err := r.Extract(&mockExtract{}); err == nil || err.Error() != "missing input/output args for extract command: {{ $.Input }} {{ $.Output }}" {
		t.Errorf("invalid error: %v", err)
	}
	r.SetAppData("a", "b", asset.Settings{Command: []string{"xyz", "{{ $.Output }}"}})
	if err := r.Extract(&mockExtract{}); err == nil || err.Error() != "missing input/output args for extract command: {{ $.Input }} {{ $.Output }}" {
		t.Errorf("invalid error: %v", err)
	}
	r.SetAppData("a", "b", asset.Settings{Command: []string{"xyz", "{{ $.Input }}"}})
	if err := r.Extract(&mockExtract{}); err == nil || err.Error() != "missing input/output args for extract command: {{ $.Input }} {{ $.Output }}" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestExtract(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	r := &asset.Resource{}
	r.File = "file"
	r.Tag = "tag"
	r.URL = "url"
	r.SetAppData("a", "testdata", asset.Settings{Command: []string{"xyz", "{{ $.Output }}", "{{ $.Input }}"}})
	if err := r.Extract(&mockExtract{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestExtractDepth(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	r := &asset.Resource{}
	r.File = "file"
	r.Tag = "tag"
	r.URL = "url"
	r.SetAppData("a", "testdata", asset.Settings{NoDepth: true, Command: []string{"xyz", "{{ $.Output }}", "{{ $.Input }}"}})
	m := &mockExtract{}
	if err := r.Extract(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m.ran) != "[xyz testdata/a.tag testdata/d970aa5.file]" {
		t.Errorf("invalid run: %v", m.ran)
	}
	r.File = "file.tar.xz"
	r.Tag = "tag2"
	r.SetAppData("a", "testdata", asset.Settings{})
	m = &mockExtract{}
	m.payload = []byte("afojea\nafofea\nlafjeha\n")
	if err := r.Extract(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", m.ran) != "[tar xf testdata/9df62cf.file.tar.xz -C testdata/a.tag2]" {
		t.Errorf("invalid run: %v", m.ran)
	}
	r.Tag = "tag4"
	r.SetAppData("a", "testdata", asset.Settings{NoDepth: true})
	m = &mockExtract{}
	m.payload = []byte("afo/jea\naojfea\nlafjeha\n")
	if err := r.Extract(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if strings.Contains(fmt.Sprintf("%v", m.ran), "strip-component") {
		t.Errorf("invalid run: %v", m.ran)
	}
	r.Tag = "tag5"
	r.SetAppData("a", "testdata", asset.Settings{})
	m = &mockExtract{}
	m.payload = []byte("afo/jea\nafo/fea\nlafjeha\n")
	if err := r.Extract(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !strings.Contains(fmt.Sprintf("%v", m.ran), "strip-component") {
		t.Errorf("invalid run: %v", m.ran)
	}
	r.File = "file.zip"
	r.Tag = "tag6"
	r.SetAppData("a", "testdata", asset.Settings{NoDepth: false})
	m = &mockExtract{}
	m.payload = []byte("afo/jea\nafo/fea\nlafjeha\n")
	if err := r.Extract(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !strings.Contains(fmt.Sprintf("%v", m.ran), " -j ") {
		t.Errorf("invalid run: %v", m.ran)
	}
	r.File = "file.zip"
	r.Tag = "tag7"
	r.SetAppData("a", "testdata", asset.Settings{NoDepth: false})
	m = &mockExtract{}
	m.payload = []byte("afo/jea\nado/fea\nlafjeha\n")
	if err := r.Extract(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if strings.Contains(fmt.Sprintf("%v", m.ran), " -j ") {
		t.Errorf("invalid run: %v", m.ran)
	}
}
