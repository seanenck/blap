// Package core maintains configuration definitions
package core

import (
	"bytes"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"github.com/seanenck/blap/internal/util"
)

const (
	disableFlag  = "disabled"
	pinFlag      = "pinned"
	redeployFlag = "redeploy"
)

type (
	// WebURL can be templated with settings from the host
	WebURL string
	// FlagSet is a simple string array to control application rules
	FlagSet []string
	// Variables define os environment variables to set
	Variables []struct {
		Key   string
		Value Resolved
	}
	// SetVariables are the variables that are set on Variables.Set and should be "unset"
	SetVariables map[string]struct {
		had   bool
		value string
	}
	// Resolved will handle env-based strings for resolution of env vars
	Resolved string
	// Filtered deals with modes that need to filters the results
	Filtered struct {
		Download string
		Filters  []string
		Sort     string
	}
	// GitHubReleaseMode are github modes operating on releases
	GitHubReleaseMode struct {
		Asset string
	}
	// GitHubBranchMode will enable a repository+branch to pull a tarball
	GitHubBranchMode struct {
		Name string
	}
	// WebMode represents web-based lookups
	WebMode struct {
		URL    WebURL
		Scrape *Filtered
	}
	// GitMode enables git-based fetches
	GitMode struct {
		Repository string
		Tagged     *Filtered
	}
	// RunMode indicates running an executable
	RunMode struct {
		Executable Resolved
		Arguments  []Resolved
		Fetch      *Filtered
	}
	// GitHubMode indicates processing of a github project for upstreams
	GitHubMode struct {
		Project string
		Release *GitHubReleaseMode
		Branch  *GitHubBranchMode
	}
	// StaticMode allows for downloading a static asset
	StaticMode struct {
		URL  WebURL
		File string
		Tag  string
	}
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Priority  int
		Flags     FlagSet
		GitHub    *GitHubMode
		Git       *GitMode
		Web       *WebMode
		Exec      *RunMode
		Static    *StaticMode
		Extract   Extraction
		Variables Variables
		ClearEnv  bool
		Setup     []Step
		Platforms []struct {
			Disable bool
			Value   Resolved
			Target  string
		}
	}
	// Step is a build process step
	Step struct {
		Directory Resolved
		Command   interface{}
		Variables Variables
		ClearEnv  bool
	}
	// Extraction handles asset extraction
	Extraction struct {
		Skip    bool
		NoDepth bool
		Command []Resolved
	}
	// CommandEnv wraps build command environment settings
	CommandEnv struct {
		Clear     bool
		Variables Variables
	}
	// Pinned are pinned names (for regex use)
	Pinned []string
	// AppSet are a name->app pair
	AppSet map[string]Application
	// GitHubSettings are overall github settings
	GitHubSettings struct {
		Token interface{}
	}
	// Connections are various endpoint settings
	Connections struct {
		GitHub   GitHubSettings
		Timeouts struct {
			Get uint
		}
	}
	// Token defines an interface for setting API/auth tokens
	Token interface {
		Env() []string
		Value() []string
	}
	// SourceType indicates if an application field contains a source
	SourceType interface {
		Is()
	}
)

// CommandEnv creates a command environment from an application
func (a Application) CommandEnv() CommandEnv {
	return CommandEnv{Clear: a.ClearEnv, Variables: a.Variables}
}

// CommandEnv creates a command environment from a step
func (s Step) CommandEnv() CommandEnv {
	return CommandEnv{Clear: s.ClearEnv, Variables: s.Variables}
}

// Is toggles on source mode
func (g GitHubMode) Is() {
}

// Is toggles on source mode
func (g GitMode) Is() {
}

// Is toggles on source mode
func (w WebMode) Is() {
}

// Is toggles on for static mode
func (s StaticMode) Is() {
}

// Is toggles running a command mode
func (r RunMode) Is() {
}

// Env will get the possible environment variables
func (g GitHubSettings) Env() []string {
	const gitHubToken = "GITHUB_TOKEN"
	return []string{"BLAP_" + gitHubToken, gitHubToken}
}

// Value will get the configured token value
func (g GitHubSettings) Value() []string {
	if util.IsNil(g.Token) {
		return nil
	}
	var vals []Resolved
	if s, ok := g.Token.(string); ok {
		vals = append(vals, Resolved(s))
	} else {
		for _, v := range g.Token.([]interface{}) {
			vals = append(vals, Resolved(v.(string)))
		}
	}
	var res []string
	for _, v := range vals {
		res = append(res, v.String())
	}
	return res
}

// Items will iterate over the available source itmes
func (a Application) Items() iter.Seq[any] {
	return func(yield func(any) bool) {
		v := reflect.ValueOf(a)
		for i := 0; i < v.NumField(); i++ {
			obj := v.Field(i).Interface()
			if _, ok := obj.(SourceType); ok {
				if !yield(obj) {
					return
				}
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
	matches := templateRegexp.FindAllString(dir, -1)
	if len(matches) > 0 {
		for _, m := range matches {
			if strings.Contains(m, ".Config.") {
				err := func() error {
					t, err := template.New("t").Parse(dir)
					if err != nil {
						return err
					}
					var buf bytes.Buffer
					if err := t.Execute(&buf, struct{ Config baseValues }{BaseTemplate}); err != nil {
						return err
					}
					dir = buf.String()
					return nil
				}()
				if err != nil {
					fmt.Fprintf(os.Stderr, "[WARNING] unable to template value %s (error: %v)", v, err)
					return dir
				}
				break
			}
		}
	}
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

// Set will set os environment variables
func (v Variables) Set() SetVariables {
	vars := make(SetVariables)
	if v == nil {
		return vars
	}
	for _, obj := range v {
		value, ok := os.LookupEnv(obj.Key)
		vars[obj.Key] = struct {
			had   bool
			value string
		}{ok, value}
		os.Setenv(obj.Key, obj.Value.String())
	}
	return vars
}

// Unset will clear/reset variables to prior state
func (v SetVariables) Unset() {
	if v == nil {
		return
	}
	for key, values := range v {
		if values.had {
			os.Setenv(key, values.value)
		} else {
			os.Unsetenv(key)
		}
	}
}

// Check will validate a flag set
func (f FlagSet) Check() error {
	l := len(f)
	switch l {
	case 0:
		return nil
	case 1, 2:
		skipped := f.Skipped()
		redeploy := f.ReDeploy()
		if l == 1 {
			if skipped || redeploy {
				return nil
			}
		} else {
			if skipped && redeploy {
				return nil
			}
		}
	}
	return fmt.Errorf("invalid flags, flag set not supported: %s", strings.Join(f, ","))
}

// ReDeploy indicates the application prefers to be redeployed
func (f FlagSet) ReDeploy() bool {
	return f.has(redeployFlag)
}

// Pin indicates if the flag means something should be pinned (not updated/disabled but not pruned)
func (f FlagSet) Pin() bool {
	return f.has(pinFlag)
}

// Skipped indicates a flag wants an actual disabled activity
func (f FlagSet) Skipped() bool {
	return f.has(pinFlag, disableFlag)
}

func (f FlagSet) has(flags ...string) bool {
	return slices.ContainsFunc(f, func(item string) bool {
		return slices.Contains(flags, item)
	})
}

// Enabled indicates if an application is enabled for use
func (a Application) Enabled() bool {
	if err := a.Flags.Check(); err != nil {
		return false
	}
	if a.Flags.Skipped() {
		return false
	}
	allowed := true
	if len(a.Platforms) > 0 {
		allowed = false
		for _, p := range a.Platforms {
			if p.Value.String() == p.Target {
				return !p.Disable
			}
		}
	}
	return allowed
}

// String returns the string of the web URL
func (w WebURL) String() string {
	return string(w)
}

// CanTemplate indicates that this type can be templated (as it can)
func (w WebURL) CanTemplate() bool {
	return true
}

// Commands will get the step commands
func (s Step) Commands() iter.Seq[[]Resolved] {
	conv := func(a []interface{}) []Resolved {
		var res []Resolved
		for _, obj := range a {
			if item, ok := obj.(string); ok {
				res = append(res, Resolved(item))
			}
		}
		return res
	}
	return func(yield func(r []Resolved) bool) {
		obj, ok := s.Command.([]interface{})
		if !ok {
			return
		}
		for idx, item := range obj {
			if idx == 0 {
				if _, ok := item.(string); ok {
					yield(conv(obj))
					return
				}
			}
			arr, ok := item.([]interface{})
			if !ok {
				continue
			}
			if !yield(conv(arr)) {
				return
			}
		}
	}
}
