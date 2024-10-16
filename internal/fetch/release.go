// Package fetch gets release/asset information from upstreams
package fetch

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/seanenck/blap/internal/asset"
)

type (
	// GitHubMode indicates processing of a github project for upstreams
	GitHubMode struct {
		Project string             `yaml:"project"`
		Release *GitHubReleaseMode `yaml:"release"`
		Branch  *GitHubBranchMode  `yaml:"branch"`
	}
	// GitHubReleaseMode are github modes operating on releases
	GitHubReleaseMode struct {
		Asset string `yaml:"asset"`
	}
	// GitHubRelease is the API definition for github release info
	GitHubRelease struct {
		Assets []struct {
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
		Tarball string `json:"tarball_url"`
		Tag     string `json:"tag_name"`
	}
)

// GitHubRelease handles GitHub-based releases
func (r *ResourceFetcher) GitHubRelease(a GitHubMode) (*asset.Resource, error) {
	up := strings.TrimSpace(a.Project)
	if up == "" {
		return nil, errors.New("github mode requires a project")
	}
	if a.Release == nil {
		return nil, errors.New("release is not properly set")
	}
	regex := a.Release.Asset
	if regex == "" {
		return nil, errors.New("github mode requires an asset filter (regex)")
	}
	tarSource := regex == tarball
	r.Context.LogDebug("getting github release: %s\n", up)
	tag, assets, err := r.latestRelease(a, tarSource)
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return nil, errors.New("no assets found")
	}
	if tag == "" {
		return nil, errors.New("assets found but no tag")
	}
	if tarSource {
		regex = ""
	}
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	var rsrc *asset.Resource
	var options []string
	for _, item := range assets {
		name := filepath.Base(item)
		options = append(options, item)
		if re.MatchString(name) {
			if rsrc != nil {
				return nil, fmt.Errorf("multiple assets matched: %s (had: %s)", item, rsrc.URL)
			}
			if tarSource {
				name = fmt.Sprintf("%s.tar.gz", name)
			}
			rsrc = &asset.Resource{URL: item, File: name, Tag: tag}
		}
	}

	if rsrc == nil {
		return nil, fmt.Errorf("unable to find asset, choices: %v", options)
	}
	return rsrc, nil
}

func (r *ResourceFetcher) latestRelease(a GitHubMode, isTarball bool) (string, []string, error) {
	release, err := fetchData[GitHubRelease](r, a.Project, "releases/latest")
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
