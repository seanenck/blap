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

func (s Settings) log(level int, msg string) {
	if s.Verbosity > level {
		fmt.Print(msg)
	}
}

// LogDebug handles debug logging
func (s Settings) LogDebug(msg string) {
	s.log(4, msg)
}

// LogInfo logs an informational message
func (s Settings) LogInfo(msg string) {
	s.log(1, msg)
}

// LogInfoSub logs a sub-step info message
func (s Settings) LogInfoSub(msg string) {
	s.LogInfo(fmt.Sprintf("  %s", msg))
}

// LogCore logs a core message
func (s Settings) LogCore(msg string) {
	s.log(0, msg)
}
