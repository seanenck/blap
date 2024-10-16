// Package fetch gets release/asset information from upstreams
package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/util"
)

const (
	tarball = "tarball"
)

type (
	// ResourceFetcher is the default fetcher for resources
	ResourceFetcher struct {
		Context   cli.Settings
		Requestor Requestable
		Execute   Executable
	}

	// GitHubErrorResponse is the error from github
	GitHubErrorResponse struct {
		Message       string `json:"message"`
		Documentation string `json:"documentation_url"`
	}

	// GitHubWrapperError indicates a download error (from github specifically)
	GitHubWrapperError struct {
		Code   int
		Status string
		Body   []byte
		URL    string
	}
	// Requestable allow for overriding GET requests
	Requestable interface {
		Do(*http.Request) (*http.Response, error)
	}
	// Executable allows for overriding command execution gathering
	Executable interface {
		Output() ([]byte, error)
	}
	// Backend is the backing way to retrieve specific resources
	Backend interface {
		GitHubRelease(GitHubMode) (*asset.Resource, error)
		Tagged(TaggedMode) (*asset.Resource, error)
		GitHubBranch(GitHubMode) (*asset.Resource, error)
	}
	// Retriever provides the means to fetch application information
	Retriever interface {
		Backend
		Download(bool, string, string) (bool, error)
		SetToken(string)
		Process(Backend, *GitHubMode, *TaggedMode) (*asset.Resource, error)
	}
)

// Process will determine the appropriate backend for processing a fetch
func (r ResourceFetcher) Process(backend Backend, gh *GitHubMode, tag *TaggedMode) (*asset.Resource, error) {
	cnt := 0
	for _, obj := range []interface{}{gh, tag} {
		if !util.IsNil(obj) {
			cnt++
			if cnt > 1 {
				return nil, errors.New("multiple modes enabled, only one allowed")
			}
		}
	}
	if gh != nil {
		if gh.Branch != nil && gh.Release != nil {
			return nil, errors.New("only one github mode is allowed")
		}
		if gh.Branch != nil {
			return backend.GitHubBranch(*gh)
		}
		if gh.Release != nil {
			return backend.GitHubRelease(*gh)
		}
		return nil, errors.New("github mode set but not configured")
	}
	if tag != nil {
		return backend.Tagged(*tag)
	}
	return nil, errors.New("unknown mode for fetch processing")
}

// Error is the interface definition for fetch errors
func (e *GitHubWrapperError) Error() string {
	components := make(map[string]string)
	for k, v := range map[string]string{
		"code":   fmt.Sprintf("%d", e.Code),
		"status": e.Status,
		"url":    e.URL,
	} {
		components[k] = v
	}
	if len(e.Body) > 0 {
		var resp GitHubErrorResponse
		err := json.Unmarshal(e.Body, &resp)
		if err == nil {
			components["message"] = resp.Message
			components["doc"] = resp.Documentation
		} else {
			components["unmarshal"] = err.Error()
		}
	}

	var msg []string
	for k, v := range components {
		if strings.TrimSpace(v) != "" {
			msg = append(msg, fmt.Sprintf("%s: %s", k, v))
		}
	}
	sort.Strings(msg)
	return strings.Join(msg, "\n")
}

func fetchData[T any](r *ResourceFetcher, ownerRepo, call string) (T, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", ownerRepo, call)
	resp, err := r.get(url)
	if err != nil {
		return *new(T), err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return *new(T), err
		}
		return *new(T), &GitHubWrapperError{
			Status: resp.Status,
			Code:   resp.StatusCode,
			Body:   body,
			URL:    url,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return *new(T), err
	}
	var obj T
	if err := json.Unmarshal(body, &obj); err != nil {
		return *new(T), err
	}
	return obj, nil
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
		resp, err := r.get(url)
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

func (r ResourceFetcher) get(url string) (*http.Response, error) {
	requestor := r.Requestor
	if r.Requestor == nil {
		requestor = &http.Client{}
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if err := r.tokenHeader(req); err != nil {
		return nil, err
	}
	return requestor.Do(req)
}

func (r *ResourceFetcher) tokenHeader(req *http.Request) error {
	var token string
	if rawToken := getToken(); rawToken != "" {
		token = rawToken
		if util.PathExists(rawToken) {
			b, err := os.ReadFile(rawToken)
			if err != nil {
				return err
			}
			token = strings.TrimSpace(string(b))
		}
	}
	if token != "" && req.URL.Scheme == "https" && req.Host == "api.github.com" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}
	return nil
}

// SetToken will set the token for the fetch (if not already set)
func (r *ResourceFetcher) SetToken(token string) {
	if getToken() == "" {
		os.Setenv(cli.BlapToken, token)
	}
}

func getToken() string {
	for _, env := range cli.TokenOptions {
		v := strings.TrimSpace(os.Getenv(env))
		if v != "" {
			return v
		}
	}
	return ""
}
