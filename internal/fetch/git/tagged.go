// Package git gets git-tagged resources
package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/filtered"
)

type taggedFilterable struct {
	filtered.Base
}

func (t taggedFilterable) Get(r fetch.Retriever, url string) ([]byte, error) {
	s, err := r.ExecuteCommand("git", "-c", "versionsort.suffix=-", "ls-remote", "--tags", "--sort=-v:refname", url)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (t taggedFilterable) Match(r []*regexp.Regexp, line string) ([]string, error) {
	for _, re := range r {
		if re.MatchString(line) {
			return []string{}, nil
		}
	}
	parts := strings.Split(line, "\t")
	if len(parts) != 2 {
		return nil, fmt.Errorf("matching version line can not be parsed: %s", line)
	}
	return []string{strings.TrimPrefix(parts[1], "refs/tags/")}, nil
}

// Tagged gets a tagged (git tag) release
func Tagged(caller fetch.Retriever, ctx fetch.Context, a core.GitMode) (*core.Resource, error) {
	f := taggedFilterable{}
	f.Base = filtered.NewBase(a.Repository, a.Tagged)
	return caller.Filtered(ctx, f)
}
