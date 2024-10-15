// Package fetch gets release/asset information from upstreams
package fetch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/seanenck/bd/internal/context"
	"github.com/seanenck/bd/internal/core"
	"github.com/seanenck/bd/internal/extract"
)

const (
	gitHubToken = "GITHUB_TOKEN"
	bdToken     = "BD_" + gitHubToken
	tarball     = "tarball"
)

// TokenOptions are the env vars for setting a github token
var TokenOptions = []string{bdToken, gitHubToken}

type (
	// ResourceFetcher is the default fetcher for resources
	ResourceFetcher struct {
		context context.Settings
	}

	// GitHubCommit is commit information from github for a repo
	GitHubCommit struct {
		Sha string `json:"sha"`
	}

	// GitHubRelease is the API definition for github release info
	GitHubRelease struct {
		Assets []struct {
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
		Tarball string `json:"tarball_url"`
		Tag     string `json:"tag_name"`
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
)

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
			components["unmarshal"] = e.Error()
		}
	}

	var msg []string
	for k, v := range components {
		if strings.TrimSpace(v) != "" {
			msg = append(msg, fmt.Sprintf("%s: %s", k, v))
		}
	}
	return strings.Join(msg, "\n")
}

func fetchRepoData[T any](ownerRepo, call string) (T, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", ownerRepo, call)
	resp, err := get(url)
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

// Branch will get an asset from a branch
func (r *ResourceFetcher) Branch(a core.BranchMode) (*extract.Asset, error) {
	if a.Branch == "" {
		return nil, errors.New("branch required for branch mode")
	}
	if a.Project == "" {
		return nil, errors.New("project required for branch mode")
	}
	commit, err := fetchRepoData[GitHubCommit](a.Project, fmt.Sprintf("commits/%s", a.Branch))
	if err != nil {
		return nil, err
	}
	tag := commit.Sha[0:7]
	return extract.NewAsset(fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", a.Project, a.Branch), fmt.Sprintf("%s-%s.tar.gz", tag, a.Branch), tag), nil
}

// Tagged gets a tagged (git tag) release
func (r *ResourceFetcher) Tagged(a core.TaggedMode) (*extract.Asset, error) {
	up := strings.TrimSpace(a.Repository)
	if up == "" {
		return nil, fmt.Errorf("no upstream for tagged mode: %v", a)
	}
	dl := strings.TrimSpace(a.Download)
	if dl == "" {
		dl = up
	}
	if len(a.Filters) == 0 {
		return nil, errors.New("application lacks filters")
	}
	out, err := exec.Command("git", "-c", "versionsort.suffix=-", "ls-remote", "--tags", "--sort=-v:refname", up).Output()
	if err != nil {
		return nil, err
	}
	var re []*regexp.Regexp
	for _, r := range a.Filters {
		c, err := regexp.Compile(r)
		if err != nil {
			return nil, err
		}
		re = append(re, c)
	}
	var tag string
	for _, line := range strings.Split(string(out), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		passed := true
		for _, r := range re {
			if r.MatchString(line) {
				passed = false
				break
			}
		}
		if passed {
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				return nil, fmt.Errorf("matching version line can not be parsed: %s", line)
			}
			tag = strings.TrimPrefix(parts[1], "refs/tags/")
			break
		}
	}
	if tag == "" {
		return nil, errors.New("no tags matched")
	}
	r.context.LogInfoSub("found tag: %s\n", tag)
	t, err := template.New("t").Parse(dl)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if err := t.Execute(&buf, struct{ Tag string }{tag}); err != nil {
		return nil, err
	}
	url := strings.TrimSpace(buf.String())
	return extract.NewAsset(url, filepath.Base(url), tag), nil
}

// GitHub gets github applications
func (r *ResourceFetcher) GitHub(a core.GitHubMode) (*extract.Asset, error) {
	up := strings.TrimSpace(a.Project)
	if up == "" {
		return nil, fmt.Errorf("github upstream is unset for: %v", a)
	}
	regex := a.Asset
	if regex == "" {
		return nil, fmt.Errorf("github mode requires an asset filter regex: %v", a)
	}
	tarSource := regex == tarball
	r.context.LogInfoSub("getting github release: %s\n", up)
	tag, assets, err := latestRelease(a, tarSource)
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return nil, errors.New("no assets found")
	}
	if tarSource {
		regex = ""
	}
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	var asset *extract.Asset
	for _, item := range assets {
		name := filepath.Base(item)
		if re.MatchString(name) {
			if asset != nil {
				return nil, fmt.Errorf("multiple assets matched: %s (had: %v)", item, asset)
			}
			if tarSource {
				name = fmt.Sprintf("%s.tar.gz", name)
			}
			asset = extract.NewAsset(item, name, tag)
		}
	}

	if asset == nil {
		return nil, fmt.Errorf("unable to find asset, choices: %v", assets)
	}
	return asset, nil
}

// SetContext will set the fetcher context for operations
func (r *ResourceFetcher) SetContext(ctx context.Settings) {
	r.context = ctx
}

// Download will download an asset
func (r *ResourceFetcher) Download(dryrun bool, url, dest string) (bool, error) {
	did := false
	if !core.PathExists(dest) {
		if dryrun {
			return true, nil
		}
		r.context.LogInfoSub("downloading asset: %s\n", url)
		resp, err := get(url)
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

func latestRelease(a core.GitHubMode, isTarball bool) (string, []string, error) {
	release, err := fetchRepoData[GitHubRelease](a.Project, "releases/latest")
	if err != nil {
		return "", nil, err
	}

	var assets []string
	if isTarball {
		assets = append(assets, release.Tarball)
	} else {
		for _, a := range release.Assets {
			assets = append(assets, a.DownloadURL)
		}
	}

	return release.Tag, assets, nil
}

func get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if err := tokenHeader(req); err != nil {
		return nil, err
	}
	client := &http.Client{}
	return client.Do(req)
}

func tokenHeader(req *http.Request) error {
	var token string
	if rawToken := getToken(); rawToken != "" {
		token = rawToken
		if core.PathExists(rawToken) {
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
		os.Setenv(bdToken, token)
	}
}

func getToken() string {
	for _, env := range TokenOptions {
		v := strings.TrimSpace(os.Getenv(env))
		if v != "" {
			return v
		}
	}
	return ""
}
