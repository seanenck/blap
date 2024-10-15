// Package context handles context for all operations
package context

import "fmt"

// InfoVerbosity is the default info level for outputs
const InfoVerbosity = 2

// Settings are the core settings
type Settings struct {
	DryRun       bool
	Purge        bool
	Applications []string
	Disabled     []string
	Verbosity    int
}

func (s Settings) log(level int, msg string, a ...any) {
	if s.Verbosity > level {
		fmt.Printf(msg, a...)
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
