// Package fetch gets release/asset information from upstreams
package fetch

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/seanenck/blap/internal/asset"
)

type (
	// TaggedMode means a repository+download is required to manage
	TaggedMode struct {
		Repository string   `yaml:"repository"`
		Download   string   `yaml:"download"`
		Filters    []string `yaml:"filters"`
	}
)

// Tagged gets a tagged (git tag) release
func (r *ResourceFetcher) Tagged(a TaggedMode) (*asset.Resource, error) {
	up := strings.TrimSpace(a.Repository)
	if up == "" {
		return nil, errors.New("no upstream for tagged mode")
	}
	dl := strings.TrimSpace(a.Download)
	if dl == "" {
		dl = up
	}
	if len(a.Filters) == 0 {
		return nil, errors.New("application lacks filters")
	}
	var re []*regexp.Regexp
	for _, r := range a.Filters {
		c, err := regexp.Compile(r)
		if err != nil {
			return nil, err
		}
		re = append(re, c)
	}
	out, err := r.executeCommand("git", "-c", "versionsort.suffix=-", "ls-remote", "--tags", "--sort=-v:refname", up)
	if err != nil {
		return nil, err
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
	r.Context.LogDebug("found tag: %s\n", tag)
	t, err := template.New("t").Parse(dl)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if err := t.Execute(&buf, struct{ Tag string }{tag}); err != nil {
		return nil, err
	}
	url := strings.TrimSpace(buf.String())
	return &asset.Resource{URL: url, File: filepath.Base(url), Tag: tag}, nil
}

func (r *ResourceFetcher) executeCommand(cmd string, args ...string) (string, error) {
	e := r.Execute
	if r.Execute == nil {
		e = exec.Command(cmd, args...)
	}
	out, err := e.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
