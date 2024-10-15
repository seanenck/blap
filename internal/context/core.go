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

// LogInfo logs an informational message
func (s Settings) LogInfo(msg string) {
	if s.Verbosity > 1 {
		fmt.Print(msg)
	}
}

// LogInfoSub logs a sub-step info message
func (s Settings) LogInfoSub(msg string) {
	s.LogInfo(fmt.Sprintf("  %s", msg))
}

// LogCore logs a core message
func (s Settings) LogCore(msg string) {
	if s.Verbosity > 0 {
		fmt.Print(msg)
	}
}
