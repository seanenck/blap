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

	"github.com/seanenck/bd/internal/core"
	"github.com/seanenck/bd/internal/extract"
	"github.com/seanenck/bd/internal/log"
)

const (
	gitHubToken = "GITHUB_TOKEN"
	bdToken     = "BD_" + gitHubToken
)

// TokenOptions are the env vars for setting a github token
var TokenOptions = []string{bdToken, gitHubToken}

type (
	// ResourceFetcher is the default fetcher for resources
	ResourceFetcher struct{}

	// GitHubRelease is the API definition for github release info
	GitHubRelease struct {
		Assets []struct {
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`

		Tag string `json:"tag_name"`
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

// Tagged gets a tagged (git tag) release
func (r ResourceFetcher) Tagged(a core.Remote) (*extract.Asset, error) {
	up := strings.TrimSpace(a.Upstream)
	if up == "" {
		return nil, fmt.Errorf("no upstream for tagged mode: %v", a)
	}
	dl := strings.TrimSpace(a.Download)
	if dl == "" {
		dl = up
	}
	if a := strings.TrimSpace(a.Asset); a != "" {
		return nil, fmt.Errorf("asset selector not allowed in tagged mode: %s", a)
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
	log.Write(fmt.Sprintf("found tag: %s\n", tag))
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
func (r ResourceFetcher) GitHub(a core.Remote) (*extract.Asset, error) {
	if dl := strings.TrimSpace(a.Download); dl != "" {
		return nil, fmt.Errorf("github mode does not support download URLs: %s", dl)
	}
	up := strings.TrimSpace(a.Upstream)
	if up == "" {
		return nil, fmt.Errorf("github upstream is unset for: %v", a)
	}
	if len(a.Filters) > 0 {
		return nil, fmt.Errorf("github mode does not support filters: %v", a.Filters)
	}
	regex := a.Asset
	if regex == "" {
		return nil, fmt.Errorf("github mode requires an asset filter regex: %v", a)
	}
	log.Write(fmt.Sprintf("getting github release: %s\n", up))
	tag, assets, err := latestRelease(a)
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return nil, errors.New("no assets found")
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
			asset = extract.NewAsset(item, name, tag)
		}
	}

	return asset, nil
}

// Download will download an asset
func (r ResourceFetcher) Download(dryrun bool, url, dest string) (bool, error) {
	did := false
	if !core.PathExists(dest) {
		if dryrun {
			return true, nil
		}
		log.Write(fmt.Sprintf("downloading asset: %s\n", url))
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

func latestRelease(a core.Remote) (string, []string, error) {
	release, err := fetchRepoData[GitHubRelease](a.Upstream, "releases/latest")
	if err != nil {
		return "", nil, err
	}

	assets := make([]string, 0, len(release.Assets))
	for _, a := range release.Assets {
		assets = append(assets, a.DownloadURL)
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
func (r ResourceFetcher) SetToken(token string) {
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
