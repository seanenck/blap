package config_test

import (
	"bytes"
	"errors"
	"fmt"
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
	"github.com/seanenck/blap/internal/purge"
	"github.com/seanenck/blap/internal/util"
)

type mockExecutor struct {
	name        string
	err         error
	rsrc        *asset.Resource
	dl          bool
	static      bool
	env         bool
	calledDo    int
	calledPurge int
	calledMulti int
}

func (m *mockExecutor) noCount() error {
	return m.expectCount(0, 0)
}

func (m *mockExecutor) expectCount(do, purge int) error {
	if m.calledDo != do*m.calledMulti || m.calledPurge != purge {
		return fmt.Errorf("do: %d (%d*%d), purge: %d (%d)", m.calledDo, do, m.calledMulti, m.calledPurge, purge)
	}
	return nil
}

func (m *mockExecutor) Purge(fxn purge.OnPurge) error {
	m.calledPurge++
	fxn()
	return m.err
}

func (m *mockExecutor) Do(ctx config.Context) error {
	m.calledDo++
	m.name = ctx.Name
	for _, e := range os.Environ() {
		if e == "ENV_KEY=some values" {
			m.env = true
		}
	}
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

func (m *mockExecutor) Run(s util.RunSettings, c string, a ...string) error {
	return m.RunCommand(c, a...)
}

func (m *mockExecutor) Changed() []string {
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
	if err := m.noCount(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cfg.Applications = make(map[string]types.Application)
	cfg.Applications["go"] = types.Application{}
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "configuration not ready" {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.noCount(); err != nil {
		t.Errorf("invalid error: %v", err)
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
	m.calledMulti = 5
	if err := m.expectCount(1, 0); err != nil {
		t.Errorf("invalid error: %v", err)
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
	if err := m.expectCount(2, 0); err != nil {
		t.Errorf("invalid error: %v", err)
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
	if err := m.expectCount(3, 0); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	str = buf.String()
	if strings.Contains(str, "xyz") || strings.Contains(str, "abc") || strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
	if m.name != "nvim" {
		t.Errorf("last app should be nvim: %s", m.name)
	}
	if !m.env {
		t.Errorf("env var not set")
	}
}

func TestProcessPurge(t *testing.T) {
	m := &mockExecutor{}
	m.rsrc = &asset.Resource{}
	m.rsrc.URL = "abc"
	m.rsrc.File = "xyz"
	m.rsrc.Tag = "tag"
	s := cli.Settings{}
	s.Purge = true
	s.DryRun = true
	s.Verbosity = cli.InfoVerbosity
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Do(config.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 1); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[abc]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Do(config.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 2); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[abc]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	s.DryRun = false
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Do(config.Context{Executor: m, Name: "sabc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 3); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[sabc]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
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
	m.err = nil
	if err := cfg.Do(config.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err == nil || err.Error() != "unexpected nil resource" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestConfigurationDoErrors(t *testing.T) {
	s := cli.Settings{}
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Do(config.Context{}); err == nil || err.Error() != "name is required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc"}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc", Fetcher: &mockExecutor{}}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc", Runner: &mockExecutor{}}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(config.Context{Name: "abc", Runner: &mockExecutor{}, Fetcher: &mockExecutor{}}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	cfg = config.Configuration{}
	if err := cfg.Do(config.Context{Fetcher: &mockExecutor{}, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err == nil || err.Error() != "configuration not setup" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestConfigurationDo(t *testing.T) {
	s := cli.Settings{}
	s.Purge = true
	f := &mockExecutor{}
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[abc]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	s.Purge = false
	s.DryRun = true
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 0 {
		t.Error("unexpected updates, no dl")
	}
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	f.dl = true
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[abc]" {
		t.Error("unexpected updates")
	}
	s.DryRun = false
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	f.rsrc = &asset.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(config.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
}

func TestConfigurationBasicProcess(t *testing.T) {
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 0 {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	str := buf.String()
	if strings.Contains(str, "DRYRUN") || !strings.Contains(str, "xyz") || !strings.Contains(str, "abc") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = true
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 0 {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	str = buf.String()
	if !strings.Contains(str, "DRYRUN") || !strings.Contains(str, "xyz") || !strings.Contains(str, "abc") {
		t.Errorf("invalid buffer: %s", str)
	}
}

func runTestIndex(do int, purging bool, afterDone, afterDryRun func(bool) error) error {
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	s.Purge = purging
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := config.Load(filepath.Join("examples", "config.yaml"), s)
	cfg.Indexing = true
	if err := cfg.Process(m, m, m); err != nil {
		return err
	}
	if len(cfg.Changed()) != 0 {
		return fmt.Errorf("invalid changed: %v", cfg.Changed())
	}
	str := buf.String()
	if strings.Contains(str, "DRYRUN") {
		return fmt.Errorf("invalid buffer: %s", str)
	}
	if !purging {
		if !strings.Contains(str, "xyz") || !strings.Contains(str, "abc") {
			return fmt.Errorf("invalid buffer: %s", str)
		}
	}
	if err := afterDone(purging); err != nil {
		return err
	}
	s.DryRun = true
	cfg, _ = config.Load(filepath.Join("examples", "config.yaml"), s)
	cfg.Indexing = true
	if err := cfg.Process(m, m, m); err != nil {
		return err
	}
	if len(cfg.Changed()) != 0 {
		return fmt.Errorf("invalid changed: %v", cfg.Changed())
	}
	m.calledMulti = 1
	if err := m.expectCount(do, 0); err != nil {
		return err
	}
	str = buf.String()
	if !strings.Contains(str, "DRYRUN") {
		return fmt.Errorf("invalid buffer: %s", str)
	}
	if !purging {
		if !strings.Contains(str, "xyz") || !strings.Contains(str, "abc") {
			return fmt.Errorf("invalid buffer: %s", str)
		}
	}
	return afterDryRun(purging)
}

func setupTestIndex(do int, setup func(), afterDone, afterDryRun func(bool) error) error {
	defer func() {
		os.RemoveAll("testdata")
	}()
	for _, b := range []bool{true, false} {
		os.RemoveAll("testdata")
		os.Mkdir("testdata", 0o755)
		setup()
		if err := runTestIndex(do, b, afterDone, afterDryRun); err != nil {
			return err
		}
	}
	return nil
}

func TestConfigurationIndexProcessNoOrEmptyFile(t *testing.T) {
	purgeIndex := filepath.Join("testdata", ".blap.purge.index")
	updateIndex := filepath.Join("testdata", ".blap.update.index")
	if err := setupTestIndex(12, func() {}, func(purge bool) error {
		for _, file := range []string{purgeIndex, updateIndex} {
			if util.PathExists(file) {
				return fmt.Errorf("run: %s should not exist (%v)", purgeIndex, purge)
			}
		}
		return nil
	}, func(purge bool) error {
		need := updateIndex
		dontNeed := purgeIndex
		if purge {
			need = purgeIndex
			dontNeed = updateIndex
		}
		for k, v := range map[string]bool{
			need:     true,
			dontNeed: false,
		} {
			if util.PathExists(k) != v {
				return fmt.Errorf("dryrun: %s unexpected, wanted %v (%v)", purgeIndex, v, purge)
			}
		}
		return nil
	}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := setupTestIndex(12, func() {
		for _, file := range []string{purgeIndex, updateIndex} {
			os.WriteFile(file, []byte("{}"), 0o644)
		}
	}, func(purge bool) error {
		need := updateIndex
		dontNeed := purgeIndex
		if !purge {
			need = purgeIndex
			dontNeed = updateIndex
		}
		for k, v := range map[string]bool{
			need:     true,
			dontNeed: false,
		} {
			if util.PathExists(k) != v {
				return fmt.Errorf("dryrun: %s unexpected, wanted %v (%v)", purgeIndex, v, purge)
			}
		}
		return nil
	}, func(purge bool) error {
		for _, file := range []string{purgeIndex, updateIndex} {
			if !util.PathExists(file) {
				return fmt.Errorf("run: %s should exist (%v)", purgeIndex, purge)
			}
		}
		return nil
	}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestConfigurationIndexProcessSet(t *testing.T) {
	purgeIndex := filepath.Join("testdata", ".blap.purge.index")
	updateIndex := filepath.Join("testdata", ".blap.update.index")
	if err := setupTestIndex(7, func() {
		for _, file := range []string{purgeIndex, updateIndex} {
			os.WriteFile(file, []byte(`{"names": ["abc", "nvim"]}`), 0o644)
		}
	}, func(purge bool) error {
		need := updateIndex
		dontNeed := purgeIndex
		if !purge {
			need = purgeIndex
			dontNeed = updateIndex
		}
		for k, v := range map[string]bool{
			need:     true,
			dontNeed: false,
		} {
			if util.PathExists(k) != v {
				return fmt.Errorf("run: %s unexpected, wanted %v (%v)", purgeIndex, v, purge)
			}
		}
		return nil
	}, func(purge bool) error {
		expect := make(map[string]string)
		expect[updateIndex] = `{"names":["abc","xyz"]}`
		expect[purgeIndex] = `{"names": ["abc", "nvim"]}`
		if purge {
			expect[updateIndex] = `{"names": ["abc", "nvim"]}`
			expect[purgeIndex] = `{"names":["abc","xyz"]}`
		}
		for k, v := range expect {
			if !util.PathExists(k) {
				return fmt.Errorf("dryrun: %s should exist (%v)", purgeIndex, purge)
			}
			b, err := os.ReadFile(k)
			if err != nil {
				return err
			}
			if string(b) != v {
				t.Errorf("invalid index: %s != %s (%v -> %s)", string(b), v, purge, k)
			}
		}
		return nil
	}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
