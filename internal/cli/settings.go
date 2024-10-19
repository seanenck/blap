// Package cli handles context for all operations
package cli

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/seanenck/blap/internal/config/types"
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
func (s Settings) ParseToken(t types.Token) (string, error) {
	for _, t := range t.Env() {
		v := strings.TrimSpace(os.Getenv(t))
		if v != "" {
			return v, nil
		}
	}
	val := t.Value().String()
	if val != "" {
		path := val
		if util.PathExists(path) {
			b, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(b)), nil
		}
	}
	return val, nil
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

func (s Settings) log(level int, msg string, a ...any) {
	settingsLock.Lock()
	defer settingsLock.Unlock()
	if s.Writer != nil && s.Verbosity > level {
		fmt.Fprintf(s.Writer, msg, a...)
	}
}

// LogDebug handles debug logging
func (s Settings) LogDebug(msg string, a ...any) {
	s.log(4, msg, a...)
}

// LogInfo logs an informational message
func (s Settings) LogInfo(msg string, a ...any) {
	s.log(1, msg, a...)
}

// LogCore logs a core message
func (s Settings) LogCore(msg string, a ...any) {
	s.log(0, msg, a...)
}
