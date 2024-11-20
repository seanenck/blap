// Package retriever gets release/asset information from upstreams
package retriever

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/git"
	"github.com/seanenck/blap/internal/fetch/github"
	"github.com/seanenck/blap/internal/fetch/web"
	"github.com/seanenck/blap/internal/util"
	"golang.org/x/mod/semver"
)

type (
	// ResourceFetcher is the default fetcher for resources
	ResourceFetcher struct {
		Context     cli.Settings
		Backend     fetch.Backend
		Connections core.Connections
		gitHubToken string
	}
)

// Process will determine the appropriate backend for processing a fetch
func (r *ResourceFetcher) Process(ctx fetch.Context, sources iter.Seq[any]) (*core.Resource, error) {
	if ctx.Name == "" {
		return nil, errors.New("name is required")
	}
	var src any
	for obj := range sources {
		if !util.IsNil(obj) {
			if src != nil {
				return nil, errors.New("multiple modes enabled, only one allowed")
			}
			src = obj
		}
	}
	switch t := src.(type) {
	case *core.GitHubMode:
		if t.Branch != nil && t.Release != nil {
			return nil, errors.New("only one github mode is allowed")
		}
		if t.Branch != nil {
			return github.Branch(r, ctx, *t)
		}
		if t.Release != nil {
			return github.Release(r, ctx, *t)
		}
		return nil, errors.New("github mode set but not configured")
	case *core.WebMode:
		if t.Scrape != nil {
			return web.Scrape(r, ctx, *t)
		}
		return nil, errors.New("unknown web mode for fetch processing")
	case *core.GitMode:
		if t.Tagged != nil {
			return git.Tagged(r, ctx, *t)
		}
		return nil, errors.New("unknown git mode for fetch processing")
	default:
		return nil, errors.New("unknown mode for fetch processing")
	}
}

// GitHubFetch performs a github fetch operations
func (r *ResourceFetcher) GitHubFetch(ownerRepo, call string, to any) error {
	if ownerRepo == "" || call == "" {
		return errors.New("owner/repo and call must be set")
	}
	if to == nil {
		return errors.New("result object must be set")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", ownerRepo, call)
	resp, err := r.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return &github.WrapperError{
			Status: resp.Status,
			Code:   resp.StatusCode,
			Body:   body,
			URL:    url,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, to)
}

// Download will download an asset
func (r *ResourceFetcher) Download(dryrun bool, url, dest string) (bool, error) {
	if url == "" || dest == "" {
		return false, errors.New("source (url) and destination (path) required")
	}
	did := false
	if !util.PathExists(dest) {
		if dryrun {
			return true, nil
		}
		r.Context.LogDebug("downloading asset: %s\n", url)
		resp, err := r.Get(url)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		if err := os.WriteFile(dest, b, 0o644); err != nil {
			return false, err
		}
		did = true
	}
	return did, nil
}

// Get performs a simple URL 'GET'
func (r ResourceFetcher) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if err := r.tokenHeader(req); err != nil {
		return nil, err
	}
	return func() (*http.Response, error) {
		if r.Backend == nil {
			cli := &http.Client{}
			if r.Connections.Timeouts.Get > 0 {
				cli.Timeout = time.Duration(r.Connections.Timeouts.Get) * time.Second
			}
			return cli.Do(req)
		}
		return r.Backend.Do(req)
	}()
}

func (r *ResourceFetcher) tokenHeader(req *http.Request) error {
	if req.URL.Scheme == "https" && req.Host == "api.github.com" {
		if r.gitHubToken == "" {
			t, err := r.Context.ParseToken(r.Connections.GitHub)
			if err != nil {
				return err
			}
			r.gitHubToken = t
		}
		if r.gitHubToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", r.gitHubToken))
		}
	}
	return nil
}

// Debug prints a debug message
func (r *ResourceFetcher) Debug(msg string, args ...any) {
	r.Context.LogDebug(msg, args...)
}

// SetConnections will configure connection information for the fetcher
func (r *ResourceFetcher) SetConnections(conn core.Connections) {
	r.Connections = conn
}

// ExecuteCommand executes an executable and args
func (r *ResourceFetcher) ExecuteCommand(cmd string, args ...string) (string, error) {
	out, err := func() ([]byte, error) {
		if r.Backend == nil {
			return exec.Command(cmd, args...).Output()
		}
		return r.Backend.Output(cmd, args...)
	}()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Filtered handles common filtered commands that have return lists of semver versions
func (r *ResourceFetcher) Filtered(ctx fetch.Context, filterable fetch.Filterable) (*core.Resource, error) {
	f := filterable.Definition()
	if f == nil {
		return nil, errors.New("filter definition is nil")
	}
	up := strings.TrimSpace(filterable.Upstream())
	if up == "" {
		return nil, errors.New("no upstream configured")
	}
	dl := strings.TrimSpace(f.Download)
	if dl == "" {
		return nil, errors.New("no download URL configured")
	}
	if len(f.Filters) == 0 {
		return nil, errors.New("filters required")
	}
	const (
		reversePrefix = "r"
		rSemVerType   = reversePrefix + "semver"
		rSortType     = reversePrefix + "sort"
		sortType      = "sort"
		semVerType    = "semver"
	)
	isSemVer := false
	isSort := false
	switch f.Sort {
	case "":
		break
	case rSortType, sortType:
		isSort = true
	case rSemVerType, semVerType:
		isSemVer = true
	default:
		return nil, fmt.Errorf("unknown sort type: %s", f.Sort)
	}
	var re []*regexp.Regexp
	for _, r := range f.Filters {
		r, err := ctx.CompileRegexp(r, nil)
		if err != nil {
			return nil, err
		}
		re = append(re, r)
	}
	b, err := filterable.Get(r, up)
	if err != nil {
		return nil, err
	}
	var options []string
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		matches, err := filterable.Match(re, t)
		if err != nil {
			return nil, err
		}
		for _, opt := range matches {
			matched := opt
			if isSemVer {
				if !strings.HasPrefix(matched, "v") {
					matched = fmt.Sprintf("v%s", matched)
				}
				if !semver.IsValid(matched) {
					r.Debug("semver found an invalid match: %s\n", matched)
					continue
				}
			}
			options = append(options, matched)
		}
	}
	if len(options) == 0 {
		return nil, errors.New("no tags found")
	}
	if isSemVer {
		semver.Sort(options)
	} else if isSort {
		sort.Strings(options)
	}
	// this seems counter to what it should be but semver/sort should be defaults to get the newest version
	// reversing should be a backup
	if f.Sort != "" && !strings.HasPrefix(f.Sort, reversePrefix) {
		slices.Reverse(options)
	}
	tag := options[0]
	r.Debug("found tag: %s\n", tag)
	tl, err := ctx.Templating(dl, &fetch.Template{Tag: fetch.Version(tag)})
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(tl)
	return &core.Resource{URL: url, File: filepath.Base(url), Tag: tag}, nil
}
