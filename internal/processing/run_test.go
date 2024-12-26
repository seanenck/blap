package processing_test

import (
	"bytes"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/logging"
	"github.com/seanenck/blap/internal/processing"
	"github.com/seanenck/blap/internal/steps"
	"github.com/seanenck/blap/internal/util"
)

type mockExecutor struct {
	name        string
	err         error
	rsrc        *core.Resource
	dl          bool
	static      bool
	env         bool
	details     string
	calledDo    int
	calledPurge int
	calledMulti int
}

func genCleanup() func() {
	os.RemoveAll("testdata")
	os.MkdirAll("testdata", 0o755)
	return func() {
		os.RemoveAll("testdata")
	}
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

func (m *mockExecutor) Purge(_ string, _ []string, fxn steps.OnPurge) error {
	m.calledPurge++
	fxn(m.details)
	return m.err
}

func (m *mockExecutor) Do(ctx processing.Context) error {
	m.calledDo++
	m.name = ctx.Name
	for _, e := range os.Environ() {
		if e == "ENV_KEY=some_values" {
			m.env = true
		}
	}
	return m.err
}

func (m *mockExecutor) Download(bool, string, string) (bool, error) {
	return m.dl, m.err
}

func (m *mockExecutor) SetConnections(core.Connections) {
}

func (m *mockExecutor) Process(fetch.Context, iter.Seq[any]) (*core.Resource, error) {
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

func (m *mockExecutor) Changed() []processing.Change {
	if m.static {
		return nil
	}
	return []processing.Change{{Name: "abc", Details: "1 details"}, {Name: "xyz", Details: "other"}}
}

func (m *mockExecutor) Debug(logging.Category, string, ...any) {
}

func (m *mockExecutor) ExecuteCommand(string, ...string) (string, error) {
	return "", nil
}

func (m *mockExecutor) GitHubFetch(string, string, any) error {
	return nil
}

func (m *mockExecutor) Get(string) (*http.Response, error) {
	return nil, nil
}

func (m *mockExecutor) Filtered(fetch.Context, fetch.Filterable) (*core.Resource, error) {
	return nil, nil
}

func TestProcessUpdate(t *testing.T) {
	cfg := processing.Configuration{}
	m := &mockExecutor{}
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.noCount(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cfg.Apps = make(map[string]core.Application)
	cfg.Apps["go"] = core.Application{}
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
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Parallelization = -1
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "parallelization must be >= 0 (have: -1)" {
		t.Errorf("invalid error: %v", err)
	}
	cfg.Parallelization = 0
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	m.calledMulti = 8
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
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(2, 0); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	str = buf.String()
	if !strings.Contains(str, "updating: abc (tag -> 1 details)") || !strings.Contains(str, "updating: xyz (tag -> other)") || !strings.Contains(str, "DRYRUN") {
		t.Errorf("invalid buffer: %s", str)
	}
	if m.name != "nvim" {
		t.Errorf("last app should be nvim: %s", m.name)
	}
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Connections.Timeouts.Get = 100
	cfg.Connections.Timeouts.All = 5
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "timeout exceeds configured 'all' settings: 100 > 5" {
		t.Errorf("invalid error: %v", err)
	}
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
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
	m.calledDo = 0
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Parallelization = 4
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.calledDo == 0 {
		t.Error("processing should have happened")
	}
}

func TestProcessPurge(t *testing.T) {
	m := &mockExecutor{}
	m.rsrc = &core.Resource{}
	m.rsrc.URL = "abc"
	m.rsrc.File = "xyz"
	m.rsrc.Tag = "tag"
	m.details = "dets"
	s := cli.Settings{}
	s.Purge = true
	s.DryRun = true
	s.Verbosity = cli.InfoVerbosity
	os.RemoveAll("testdata")
	os.MkdirAll(filepath.Join("testdata", "abc"), 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err == nil || err.Error() != "unable to purge when current release assets are not deployed: testdata/abc/9e6ff33.xyz" {
		t.Errorf("invalid error: %v", err)
	}
	os.MkdirAll(filepath.Join("testdata", "abc", "9e6ff33.abc.tag"), 0o755)
	os.MkdirAll(filepath.Join("testdata", "abc", "9e6ff33.xyz"), 0o755)
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 1); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[{abc dets}]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	m.details = ""
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 2); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[{abc }]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	s.DryRun = true
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{Executor: m, Name: "sabc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 2); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	s.DryRun = false
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{Executor: m, Name: "sabc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 2); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	os.Mkdir(filepath.Join("testdata", "sabc", "9e6ff33.sabc.tag"), 0o755)
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := m.expectCount(0, 3); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[{abc }]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
}

func TestProcessDoError(t *testing.T) {
	defer genCleanup()()
	m := &mockExecutor{}
	m.err = errors.New("ERROR")
	s := cli.Settings{}
	s.Purge = true
	s.Verbosity = cli.InfoVerbosity
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err == nil || !strings.Contains(err.Error(), "error: ERROR") {
		t.Errorf("invalid error: %v", err)
	}
	m.err = nil
	if err := cfg.Do(processing.Context{Executor: m, Name: "abc", Fetcher: m, Runner: m}); err == nil || err.Error() != "unexpected nil resource" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestConfigurationDoErrors(t *testing.T) {
	s := cli.Settings{}
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Do(processing.Context{}); err == nil || err.Error() != "name is required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(processing.Context{Name: "abc"}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(processing.Context{Name: "abc", Fetcher: &mockExecutor{}}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(processing.Context{Name: "abc", Runner: &mockExecutor{}}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	if err := cfg.Do(processing.Context{Name: "abc", Runner: &mockExecutor{}, Fetcher: &mockExecutor{}}); err == nil || err.Error() != "fetcher, runner, and executor must be set" {
		t.Errorf("invalid error: %v", err)
	}
	cfg = processing.Configuration{}
	if err := cfg.Do(processing.Context{Fetcher: &mockExecutor{}, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err == nil || err.Error() != "configuration not setup" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestConfigurationDo(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	s := cli.Settings{}
	s.Purge = true
	f := &mockExecutor{}
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	os.Mkdir(filepath.Join("testdata", "abc", "2f6fd3b.abc.123"), 0o755)
	if err := cfg.Do(processing.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[]" {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	s.Purge = false
	s.DryRun = true
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(processing.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 0 {
		t.Error("unexpected updates, no dl")
	}
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	f.dl = true
	if err := cfg.Do(processing.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", cfg.Changed()) != "[{abc 123}]" {
		t.Errorf("unexpected updates: %v", cfg.Changed())
	}
	s.DryRun = false
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(processing.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
	s.Verbosity = 100
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	app := core.Application{}
	app.Setup = append(app.Setup, core.Step{})
	app.Extract.Skip = true
	if err := cfg.Do(processing.Context{Application: app, Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
	str := buf.String()
	if !strings.Contains(str, "no extraction") || !strings.Contains(str, "steps set for") {
		t.Error("should not extract")
	}
}

func TestReDeploy(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	s := cli.Settings{}
	f := &mockExecutor{}
	f.dl = true
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(processing.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
	s.Verbosity = 100
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(processing.Context{Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
	if s := buf.String(); !strings.Contains(s, "deployed:") {
		t.Errorf("invalid buffer: %s", s)
	}
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(processing.Context{Application: core.Application{Flags: []string{"redeploy"}}, Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
	if s := buf.String(); strings.Contains(s, "deployed:") {
		t.Errorf("invalid buffer: %s", s)
	}
	buf = bytes.Buffer{}
	s.Writer = &buf
	s.ReDeploy = true
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	f.rsrc = &core.Resource{File: "xyz.tar.xz", URL: "xxx", Tag: "123"}
	if err := cfg.Do(processing.Context{Application: core.Application{}, Fetcher: f, Name: "abc", Runner: &mockExecutor{}, Executor: &mockExecutor{}}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 1 {
		t.Error("unexpected updates")
	}
	if s := buf.String(); strings.Contains(s, "deployed:") {
		t.Errorf("invalid buffer: %s", s)
	}
}

func TestConfigurationBasicProcess(t *testing.T) {
	defer genCleanup()()
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 0 {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	str := buf.String()
	if strings.Contains(str, "DRYRUN") || !strings.Contains(str, "updating: xyz (tag -> other)") || !strings.Contains(str, "updating: abc (tag -> 1 details)") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = true
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(cfg.Changed()) != 0 {
		t.Errorf("invalid changed: %v", cfg.Changed())
	}
	str = buf.String()
	if !strings.Contains(str, "DRYRUN") || !strings.Contains(str, "updating: xyz (tag -> other)") || !strings.Contains(str, "updating: abc (tag -> 1 details)") {
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
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Indexing.Enabled = true
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
	messageOne := "purging: abc (directory -> 1 details)"
	messageTwo := "purging: xyz (directory -> other)"
	if !purging {
		messageOne = "updating: abc (tag -> 1 details)"
		messageTwo = "updating: xyz (tag -> other)"
	}
	if !strings.Contains(str, messageOne) || !strings.Contains(str, messageTwo) {
		return fmt.Errorf("invalid buffer: %s", str)
	}
	if err := afterDone(purging); err != nil {
		return err
	}
	s.DryRun = true
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Indexing.Enabled = true
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
	if !strings.Contains(str, messageTwo) || !strings.Contains(str, messageOne) {
		return fmt.Errorf("invalid buffer: %s", str)
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
	if err := setupTestIndex(18, func() {}, func(purge bool) error {
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
	if err := setupTestIndex(18, func() {
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
	if err := setupTestIndex(10, func() {
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

func TestCleanDirs(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	s.Purge = true
	s.CleanDirs = true
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s.DryRun = true
	buf = bytes.Buffer{}
	s.Verbosity = 100
	s.Writer = &buf
	os.Mkdir(filepath.Join("testdata", "zzzzzzz"), 0o755)
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "nvim"), 0o755)
	os.WriteFile(filepath.Join("testdata", "test.toml"), []byte{}, 0o644)
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	dirs, _ := os.ReadDir("testdata")
	if len(dirs) != 4 {
		t.Errorf("invalid dirs: %v", dirs)
	}
	str := buf.String()
	if !strings.Contains(str, "DRYRUN") || !strings.Contains(str, "removing directory: abc") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = false
	buf = bytes.Buffer{}
	s.Writer = &buf
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	dirs, _ = os.ReadDir("testdata")
	if len(dirs) != 3 {
		t.Errorf("invalid dirs: %v", dirs)
	}
	str = buf.String()
	if strings.Contains(str, "DRYRUN") || !strings.Contains(str, "removing directory: abc") {
		t.Errorf("invalid buffer: %s", str)
	}
}

func TestCleanDirsIndexed(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	s.Purge = true
	s.CleanDirs = true
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s.DryRun = true
	buf = bytes.Buffer{}
	s.Verbosity = 100
	s.Writer = &buf
	os.Mkdir(filepath.Join("testdata", "zzzzzzz"), 0o755)
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "nvim"), 0o755)
	os.WriteFile(filepath.Join("testdata", "test.toml"), []byte{}, 0o644)
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Indexing.Enabled = true
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	dirs, _ := os.ReadDir("testdata")
	if len(dirs) != 5 {
		t.Errorf("invalid dirs: %v", dirs)
	}
	b, _ := os.ReadFile(filepath.Join("testdata", ".blap.purge.index"))
	if strings.Contains(string(b), `"dirs:["abc"]`) {
		t.Errorf("invalid index: %s", string(b))
	}
	str := buf.String()
	if !strings.Contains(str, "DRYRUN") || !strings.Contains(str, "removing directory: abc") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = false
	buf = bytes.Buffer{}
	s.Writer = &buf
	os.Mkdir(filepath.Join("testdata", "123"), 0o755)
	cfg, _ = processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Indexing.Enabled = true
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	dirs, _ = os.ReadDir("testdata")
	if len(dirs) != 4 {
		t.Errorf("invalid dirs: %v", dirs)
	}
	str = buf.String()
	if strings.Contains(str, "DRYRUN") || !strings.Contains(str, "removing directory: abc") {
		t.Errorf("invalid buffer: %s", str)
	}
}

func TestStrictIndexed(t *testing.T) {
	defer genCleanup()()
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	s.Purge = true
	s.CleanDirs = true
	var buf bytes.Buffer
	s.Writer = &buf
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	cfg.Indexing.Enabled = true
	cfg.Indexing.Strict = true
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "index not found: testdata/.blap.purge.index (strict mode)" {
		t.Errorf("invalid error: %v", err)
	}
	cfg.Indexing.Enabled = false
	if err := cfg.Process(m, m, m); err == nil || err.Error() != "can not enable strict indexing without indexing enabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestLogging(t *testing.T) {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	defer func() {
		os.RemoveAll("testdata")
	}()
	m := &mockExecutor{}
	s := cli.Settings{}
	s.Verbosity = cli.InfoVerbosity
	s.Purge = true
	s.CleanDirs = true
	b, _ := os.ReadFile(filepath.Join("examples", "config.toml"))
	logFile := filepath.Join("testdata", "blap.log")
	data := strings.ReplaceAll(string(b), "file = \"\"", fmt.Sprintf("file = \"%s\"", logFile))
	to := filepath.Join("testdata", "config.toml")
	os.WriteFile(to, []byte(data), 0o644)
	cfg, _ := processing.Load(to, s)
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	b, _ = os.ReadFile(filepath.Join("testdata", "blap.log"))
	str := string(b)
	if strings.Contains(str, "DRYRUN") || !strings.Contains(str, "purging") {
		t.Errorf("invalid buffer: %s", str)
	}
	s.DryRun = true
	s.Verbosity = 100
	os.Mkdir(filepath.Join("testdata", "zzzzzzz"), 0o755)
	os.Mkdir(filepath.Join("testdata", "abc"), 0o755)
	os.Mkdir(filepath.Join("testdata", "nvim"), 0o755)
	os.WriteFile(filepath.Join("testdata", "test.toml"), []byte("{}"), 0o644)
	cfg, _ = processing.Load(to, s)
	cfg.Indexing.Enabled = true
	if err := cfg.Process(m, m, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	b, _ = os.ReadFile(filepath.Join("testdata", "blap.log"))
	str = string(b)
	if !strings.Contains(str, "DRYRUN") || !strings.Contains(str, "purging") {
		t.Errorf("invalid buffer: %s", str)
	}
}

func TestLock(t *testing.T) {
	defer genCleanup()()
	f := filepath.Join("testdata", "lock")
	os.WriteFile(f, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
	s := cli.Settings{}
	cfg, _ := processing.Load(filepath.Join("examples", "config.toml"), s)
	if err := cfg.Lock(f); err == nil || !strings.HasPrefix(err.Error(), "instance already running, has lock: testdata/lock (pid: ") {
		t.Errorf("invalid error: %v", err)
	}
	os.Remove(f)
	if err := cfg.Lock(f); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
