// Package cli handles context for all operations
package cli

import (
	"fmt"
	"io"
	"regexp"

	"github.com/seanenck/blap/internal/util"
)

// InfoVerbosity is the default info level for outputs
const InfoVerbosity = 2

// Settings are the core settings
type Settings struct {
	DryRun bool
	Purge  bool
	Writer io.Writer
	filter struct {
		has    bool
		negate bool
		regex  *regexp.Regexp
	}
	Verbosity int
	Resolves  map[string]string
}

// FilterApplications indicates if the
func (s Settings) FilterApplications() bool {
	return s.filter.has
}

// Resolve will cache resolve paths
func (s Settings) Resolve(dir string) string {
	if s.Resolves != nil {
		has, ok := s.Resolves[dir]
		if ok {
			return has
		}
	}
	res := util.ResolveDirectory(dir)
	if s.Resolves != nil {
		s.Resolves[dir] = res
	}
	return res
}

// AllowApplication indicates if an application is allowed
func (s Settings) AllowApplication(input string) bool {
	if !s.filter.has {
		return true
	}
	m := s.filter.regex.MatchString(input)
	if s.filter.negate {
		m = !m
	}
	return m
}

// CompileApplicationFilter will compile the necessary app filter
func (s *Settings) CompileApplicationFilter(filter string, negate bool) error {
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
