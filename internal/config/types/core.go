// Package types maintains configuration definitions
package types

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type (
	Variables struct {
		Vars map[string]string `yaml:"values"`
		had  map[string]string
	}
	// Resolved will handle env-based strings for resolution of env vars
	Resolved string
	// GitTaggedMode means a repository+download is required to manage
	GitTaggedMode struct {
		Download string   `yaml:"download"`
		Filters  []string `yaml:"filters"`
	}
	// GitHubReleaseMode are github modes operating on releases
	GitHubReleaseMode struct {
		Asset string `yaml:"asset"`
	}
	// GitHubBranchMode will enable a repository+branch to pull a tarball
	GitHubBranchMode struct {
		Name string `yaml:"name"`
	}
	// GitMode enables git-based fetches
	GitMode struct {
		Repository string         `yaml:"repository"`
		Tagged     *GitTaggedMode `yaml:"tagged"`
	}
	// GitHubMode indicates processing of a github project for upstreams
	GitHubMode struct {
		Project string             `yaml:"project"`
		Release *GitHubReleaseMode `yaml:"release"`
		Branch  *GitHubBranchMode  `yaml:"branch"`
	}
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Priority int        `yaml:"priority"`
		Disable  bool       `yaml:"disable"`
		Source   Source     `yaml:"source"`
		Extract  Extraction `yaml:"extract"`
		Commands struct {
			Environment CommandEnvironment `yaml:"environment"`
			Steps       []Step             `yaml:"steps"`
		} `yaml:"commands"`
	}
	// Step is a build process step
	Step struct {
		Directory   Resolved           `yaml:"directory"`
		Command     []Resolved         `yaml:"command"`
		Environment CommandEnvironment `yaml:"environment"`
	}
	// CommandEnvironment are environment configuration settings for build steps
	CommandEnvironment struct {
		Variables Variables `yaml:"variables"`
		Clear     bool      `yaml:"clear"`
	}
	// Extraction handles asset extraction
	Extraction struct {
		NoDepth bool       `yaml:"nodepth"`
		Command []Resolved `yaml:"command"`
	}
	// Source are the available source options
	Source struct {
		GitHub *GitHubMode `yaml:"github"`
		Git    *GitMode    `yaml:"git"`
	}
	// Pinned are pinned names (for regex use)
	Pinned []string
	// AppSet are a name->app pair
	AppSet map[string]Application
	// GitHubSettings are overall github settings
	GitHubSettings struct {
		Token Resolved `yaml:"token"`
	}
	// Connections are various endpoint settings
	Connections struct {
		GitHub GitHubSettings `yaml:"github"`
	}
	// Token defines an interface for setting API/auth tokens
	Token interface {
		Env() []string
		Value() Resolved
	}
)

// Env will get the possible environment variables
func (g GitHubSettings) Env() []string {
	const gitHubToken = "GITHUB_TOKEN"
	return []string{"BLAP_" + gitHubToken, gitHubToken}
}

// Value will get the configured token value
func (g GitHubSettings) Value() Resolved {
	return g.Token
}

// Items will iterate over the available source itmes
func (s Source) Items() iter.Seq[any] {
	return func(yield func(any) bool) {
		v := reflect.ValueOf(s)
		for i := 0; i < v.NumField(); i++ {
			if !yield(v.Field(i).Interface()) {
				return
			}
		}
	}
}

// String will resolve ~/ and basic env vars
func (r Resolved) String() string {
	v := string(r)
	if v == "" {
		return v
	}
	dir := os.Expand(v, os.Getenv)
	isHome := fmt.Sprintf("~%c", os.PathSeparator)
	if !strings.HasPrefix(dir, isHome) {
		return dir
	}
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return dir
	}
	return filepath.Join(h, strings.TrimPrefix(dir, isHome))
}

func (v *Variables) Set() {
	if v.Vars == nil {
		return
	}
	if v.had == nil {
		v.had = make(map[string]string)
	}
	for key, val := range v.Vars {
		had, ok := os.LookupEnv(key)
		if ok {
			v.had[key] = had
		}
		os.Setenv(key, val)
	}
}

func (v *Variables) Unset() {
	if v.Vars == nil || v.had == nil {
		return
	}
	for key := range v.Vars {
		if was, ok := v.had[key]; ok {
			os.Setenv(key, was)
		} else {
			os.Unsetenv(key)
		}
	}
}
