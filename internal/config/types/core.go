// Package types maintains configuration definitions
package types

import (
	"iter"
	"os"
	"reflect"
	"strings"

	"github.com/seanenck/blap/internal/util"
)

type (
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
		Build    struct {
			Environment BuildEnvironment `yaml:"environment"`
			Steps       []Step           `yaml:"steps"`
		} `yaml:"build"`
		Deploy struct {
			Artifacts []Artifact `yaml:"artifacts"`
		} `yaml:"deploy"`
	}
	// Artifact are definitions of what to deploy
	Artifact struct {
		Files       []string `yaml:"files"`
		Destination string   `yaml:"destination"`
	}
	// Step is a build process step
	Step struct {
		Directory   string           `yaml:"directory"`
		Command     []string         `yaml:"command"`
		Environment BuildEnvironment `yaml:"environment"`
	}
	// BuildEnvironment are environment configuration settings for build steps
	BuildEnvironment struct {
		Clear  bool     `yaml:"clear"`
		Values []string `yaml:"values"`
	}
	// Extraction handles asset extraction
	Extraction struct {
		NoDepth bool     `yaml:"nodepth"`
		Command []string `yaml:"command"`
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
		Token string `yaml:"token"`
	}
	// Connections are various endpoint settings
	Connections struct {
		GitHub GitHubSettings `yaml:"github"`
	}
	// Token defines an interface for setting API/auth tokens
	Token interface {
		Env() []string
		Value() string
	}
)

// Env will get the possible environment variables
func (g GitHubSettings) Env() []string {
	const gitHubToken = "GITHUB_TOKEN"
	return []string{"BLAP_" + gitHubToken, gitHubToken}
}

// Value will get the configured token value
func (g GitHubSettings) Value() string {
	return g.Token
}

// ParseToken will handle determine the appropriate token to use
func ParseToken(t Token) (string, error) {
	for _, t := range t.Env() {
		v := strings.TrimSpace(os.Getenv(t))
		if v != "" {
			return v, nil
		}
	}
	val := t.Value()
	if val != "" {
		if util.PathExists(val) {
			b, err := os.ReadFile(val)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(b)), nil
		}
	}
	return val, nil
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
