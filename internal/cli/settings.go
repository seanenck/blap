// Package cli handles context for all operations
package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/logging"
	"github.com/seanenck/blap/internal/util"
)

// InfoVerbosity is the default info level for outputs
const InfoVerbosity = 2

var settingsLock = &sync.Mutex{}

// Settings are the core settings
type Settings struct {
	DryRun  bool
	Purge   bool
	Writer  io.Writer
	Include *regexp.Regexp
	filter  struct {
		has    bool
		negate bool
		regex  *regexp.Regexp
	}
	Verbosity int
	CleanDirs bool
}

// FilterApplications indicates if the
func (s Settings) FilterApplications() bool {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	return s.filter.has
}

// AllowApplication indicates if an application is allowed
func (s Settings) AllowApplication(input string) bool {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	if !s.filter.has {
		return true
	}
	m := s.filter.regex.MatchString(input)
	if s.filter.negate {
		m = !m
	}
	return m
}

// ParseToken will handle determine the appropriate token to use
func (s Settings) ParseToken(t core.Token) (string, error) {
	for _, t := range t.Env() {
		v := strings.TrimSpace(os.Getenv(t))
		if v != "" {
			return v, nil
		}
	}
	token, err := func() (string, error) {
		raw := t.Value()
		var cmd string
		var args []string
		switch len(raw) {
		case 0:
			return "", nil
		case 1:
			p := raw[0]
			if !util.PathExists(p) {
				return p, nil
			}
			info, err := os.Stat(p)
			if err != nil {
				return "", err
			}
			if info.Mode()&0o111 != 0 {
				cmd = p
				break
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return "", err
			}
			return string(b), err
		default:
			cmd = raw[0]
			args = raw[1:]
		}
		b, err := exec.Command(cmd, args...).Output()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(token), nil
}

// CompileApplicationFilter will compile the necessary app filter
func (s *Settings) CompileApplicationFilter(filter string, negate bool) error {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	if filter == "" {
		return nil
	}
	s.filter.has = true
	s.filter.negate = negate
	re, err := regexp.Compile(filter)
	if err != nil {
		return err
	}
	s.filter.regex = re
	return nil
}

func (s Settings) log(level int, cat logging.Category, msg string, a ...any) {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	if s.Writer != nil && s.Verbosity > level {
		fmt.Fprintf(s.Writer, "[%s] %s", cat, fmt.Sprintf(msg, a...))
	}
}

// LogDebug handles debug logging
func (s Settings) LogDebug(cat logging.Category, msg string, a ...any) {
	s.log(4, cat, msg, a...)
}

// LogCore logs a core message
func (s Settings) LogCore(cat logging.Category, msg string, a ...any) {
	s.log(0, cat, msg, a...)
}
