// Package github gets release/asset information from upstreams
package github

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/logging"
)

// Release handles GitHub-based releases
func Release(caller fetch.Retriever, ctx fetch.Context, a core.GitHubMode) (*core.Resource, error) {
	up := strings.TrimSpace(a.Project)
	if up == "" {
		return nil, errors.New("release mode requires a project")
	}
	if a.Release == nil {
		return nil, errors.New("release is not properly set")
	}
	regex := a.Release.Asset
	if regex == "" {
		return nil, errors.New("release mode requires an asset filter (regex)")
	}
	tarSource := regex == "tarball"
	caller.Debug(logging.GitHubCategory, "getting github release: %s\n", up)
	tag, assets, err := latestRelease(caller, a, tarSource)
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

	re, err := ctx.CompileRegexp(regex, &fetch.Template{Tag: core.Version(tag)})
	if err != nil {
		return nil, err
	}
	var rsrc *core.Resource
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
			rsrc = &core.Resource{URL: item, File: name, Tag: tag}
		}
	}

	if rsrc == nil {
		return nil, fmt.Errorf("unable to find asset, choices: %v", options)
	}
	return rsrc, nil
}

func latestRelease(caller fetch.Retriever, a core.GitHubMode, isTarball bool) (string, []string, error) {
	type (
		Release struct {
			Assets []struct {
				DownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
			Tarball string `json:"tarball_url"`
			Tag     string `json:"tag_name"`
		}
	)
	release := Release{}
	if err := caller.GitHubFetch(a.Project, "releases/latest", &release); err != nil {
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
