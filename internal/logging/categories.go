// Package logging handles log helper
package logging

// Category are logging category definitions
type Category string

const (
	// BuildCategory indicates outputs for build steps
	BuildCategory = "build"
	// FetchCategory is for resource fetching/downloading
	FetchCategory = "fetch"
	// ConfigCategory is for anything related to config loading/parsing
	ConfigCategory = "config"
	// SelfCategory are internal/self messages
	SelfCategory = "internal"
	// IndexCategory are for index operations
	IndexCategory = "index"
	// ProcessCategory are for (overall) processing steps
	ProcessCategory = "process"
	// ExtractCategory are for extraction items
	ExtractCategory = "extract"
	// FilteringCategory are for fetch-filter retrievals
	FilteringCategory = "filtering"
	// GitHubCategory are for github-based logging needs
	GitHubCategory = "github"
)
