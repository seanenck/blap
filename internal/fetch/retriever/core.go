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

	"github.com/seanenck/blap/internal/asset"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/config/types"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/git"
	"github.com/seanenck/blap/internal/fetch/github"
	"github.com/seanenck/blap/internal/util"
)

type (
	// ResourceFetcher is the default fetcher for resources
	ResourceFetcher struct {
		Context     cli.Settings
		Backend     fetch.Backend
		Connections types.Connections
	}
)

// Process will determine the appropriate backend for processing a fetch
func (r *ResourceFetcher) Process(ctx fetch.Context, sources iter.Seq[any]) (*asset.Resource, error) {
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
	case *types.GitHubMode:
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
	case *types.GitMode:
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
	resp, err := r.get(url)
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
			return cli.Do(req)
		}
		return r.Backend.Do(req)
	}()
}

func (r *ResourceFetcher) tokenHeader(req *http.Request) error {
	token, err := types.ParseToken(r.Connections.GitHub)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}
	return nil
}

// Debug prints a debug message
func (r *ResourceFetcher) Debug(msg string, args ...any) {
	r.Context.LogDebug(msg, args...)
}

// SetConnections will configure connection information for the fetcher
func (r *ResourceFetcher) SetConnections(conn types.Connections) {
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
