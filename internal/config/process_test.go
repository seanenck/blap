package config_test

import (
	"bytes"
	"errors"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config"
	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/util"
)

type mockExecutor struct {
	mode   string
	name   string
	err    error
	rsrc   *asset.Resource
	dl     bool
	purged bool
	static bool
}

func (m *mockExecutor) Purge() (bool, error) {
	m.mode = "purge"
	return m.purged, m.err
}

func (m *mockExecutor) Do(ctx config.Context) error {
	m.mode = "do"
	m.name = ctx.Name
	return m.err
}

func (m *mockExecutor) Download(bool, string, string) (bool, error) {
	return m.dl, m.err
}

func (m *mockExecutor) SetConnections(types.Connections) {
}

func (m *mockExecutor) Process(fetch.Context, iter.Seq[any]) (*asset.Resource, error) {
	return m.rsrc, m.err
}

func (m *mockExecutor) RunCommand(string, ...string) error {
	return m.err
}

func (m *mockExecutor) Output(string, ...string) ([]byte, error) {
	return nil, m.err
}

func (m *mockExecutor) Run(util.RunSettings, string, ...string) error {
	return m.err
}

func (m *mockExecutor) Updated() []string {
	if m.static {
		return nil
	}
	return []string{"abc", "xyz"}
}

func (m *mockExecutor) Debug(string, ...any) {
}

func (m *mockExecutor) ExecuteCommand(string, ...string) (string, error) {
	return "", nil
}

func (m *mockExecutor) GitHubFetch(string, string, any) error {
	return nil
}

func TestProcessUpdate(t *testing.T) {
	cfg := config.Configuration{}
	m := &mockExecutor{}
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	cfg.Applications = make(map[string]types.Application)
	cfg.Applications["go"] = types.Application{}
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "configuration not ready" {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	cfg.Parallelization = -1
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "parallelization must be >= 0 (have: -1)" {
		t.Errorf("invalid error: %v", err)
	}
	cfg.Parallelization = 4
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "do" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	str := buf.String()
	if !strings.Contains(str, "xyz") || !strings.Contains(str, "abc") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = true
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "do" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	str = buf.String()
	if !strings.Contains(str, "xyz") || !strings.Contains(str, "abc") || !strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
	if m.name != "nvim" {
		t.Errorf("last app should be nvim: %s", m.name)
	}
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	m.static = true
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "do" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	str = buf.String()
	if strings.Contains(str, "xyz") || strings.Contains(str, "abc") || strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
	if m.name != "nvim" {
		t.Errorf("last app should be nvim: %s", m.name)
	}
}

func TestProcessPurge(t *testing.T) {
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Purge = true
	s.DryRun = true
	s.Verbosity = cli.InfoVerbosity
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "purge" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	str := buf.String()
	if strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
	m.purged = true
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "purge" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	str = buf.String()
	if !strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = false
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.mode != "purge" {
		t.Errorf("invalid mode: %s", m.mode)
	}
	str = buf.String()
	if strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
}

func TestProcessDoError(t *testing.T) {
	m := &mockExecutor{}
	m.err = errors.New("ERROR")
	s := cli.Settings{}
	s.Purge = true
	s.Verbosity = cli.InfoVerbosity
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err == nil || !strings.Contains(err.Error(), "error: ERROR") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestConfigurationDo(t *testing.T) {
	s := cli.Settings{}
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Do(config.Context{}); err == nil || err.Error() != "name is required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc"}); err == nil || err.Error() != "fetcher and runner must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc", Fetcher: &mockExecutor{}}); err == nil || err.Error() != "fetcher and runner must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc", Runner: &mockExecutor{}}); err == nil || err.Error() != "fetcher and runner must be set" {
		t.Errorf("invalid error: %v", err)
	}
	cfg = config.Configuration{}
	if err := cfg.Do(config.Context{Fetcher: &mockExecutor{}, Name: "abc", Runner: &mockExecutor{}}); err == nil || err.Error() != "configuration not setup" {
		t.Errorf("invalid error: %v", err)
	}
	s.Purge = true
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f := &mockExecutor{}
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Updated()) != 0 {
		t.Error("unexpected updates, purge")
	}
	s.Purge = false
	s.DryRun = true
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Updated()) != 0 {
		t.Error("unexpected updates, no dl")
	}
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	f.dl = true
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Updated()) != 1 {
		t.Error("unexpected updates")
	}
	s.DryRun = false
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Updated()) != 1 {
		t.Error("unexpected updates")
	}
}
