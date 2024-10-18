// Package cli handles environment variables
package cli

import (
	"os"
	"path/filepath"
	"sort"
)

const (
	// ConfigFileEnv is the environment variable for config file override
	ConfigFileEnv = "BLAP_CONFIG_FILE"
)

// DefaultConfigs is the list of options for config files
func DefaultConfigs() []string {
	var opts []string
	for k, v := range map[string]string{
		"HOME":            ".config",
		"XDG_CONFIG_HOME": "",
	} {
		p := os.Getenv(k)
		if p == "" {
			continue
		}
		if v != "" {
			p = filepath.Join(p, v)
		}
		opts = append(opts, configPath(p))
	}
	sort.Strings(opts)
	return opts
}
