// Package git gets git-tagged resources
package git

import (
	"fmt"
	"strings"

	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/fetch"
	"github.com/seanenck/blap/internal/fetch/filtered"
)

type taggedFilterable struct{}

func (t taggedFilterable) Get(r fetch.Retriever, url string) ([]byte, error) {
	s, err := r.ExecuteCommand("git", "-c", "versionsort.suffix=-", "ls-remote", "--tags", "--sort=-v:refname", url)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (t taggedFilterable) NewLine(line string) (string, error) {
	parts := strings.Split(line, "\t")
	if len(parts) != 2 {
		return "", fmt.Errorf("matching version line can not be parsed: %s", line)
	}
	return strings.TrimPrefix(parts[1], "refs/tags/"), nil
}

func (t taggedFilterable) Arguments() []string {
	return nil
}

// Tagged gets a tagged (git tag) release
func Tagged(caller fetch.Retriever, ctx fetch.Context, a core.GitMode) (*core.Resource, error) {
	b, err := filtered.NewBase(a.Repository, a.Tagged, taggedFilterable{})
	if err != nil {
		return nil, err
	}
	return b.Get(caller, ctx)
}
