// Package processing loads yaml configs
package processing

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
)

func doDecode[T any](in string, o T) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())
	if err := decoder.Decode(o); err != nil {
		return fmt.Errorf("file: %s -> %v", in, err)
	}
	return nil
}

// Load will initialize the configuration from a file
func Load(input string, context cli.Settings) (Configuration, error) {
	c := Configuration{}
	c.handler = &processHandler{}
	c.context = context
	c.Applications = make(map[string]core.Application)
	if err := doDecode(input, &c); err != nil {
		return c, err
	}
	c.logFile = c.Log.String()
	c.dir = c.Directory.String()
	if len(c.Include) > 0 {
		hasIncludefilter := context.Include != nil
		var including []string
		for _, i := range c.Include {
			r := i.String()
			res := []string{r}
			c.context.LogDebug("including: %s\n", i)
			if strings.Contains(r, "*") {
				globbed, err := filepath.Glob(r)
				if err != nil {
					return c, err
				}
				res = globbed
			}
			including = append(including, res...)
		}
		for _, include := range including {
			c.context.LogDebug("loading included: %s\n", include)
			if hasIncludefilter {
				if !context.Include.MatchString(include) {
					c.context.LogDebug("file does not match include filter\n")
					continue
				}
			}
			type included struct {
				Applications core.AppSet `yaml:"applications"`
				Disable      bool        `yaml:"disable"`
				Pinned       core.Pinned `yaml:"pinned"`
			}
			var apps included
			if err := doDecode(include, &apps); err != nil {
				return c, err
			}
			if apps.Disable {
				continue
			}
			c.Pinned = append(c.Pinned, apps.Pinned...)
			for k, v := range apps.Applications {
				if _, ok := c.Applications[k]; ok {
					return c, fmt.Errorf("%s is overwritten by config: %s", k, include)
				}
				c.Applications[k] = v
			}
		}
	}
	canFilter := context.FilterApplications()
	sub := make(map[string]core.Application)
	for n, a := range c.Applications {
		if a.Disable {
			continue
		}
		allowed := true
		if canFilter {
			allowed = context.AllowApplication(n)
		}
		if allowed {
			sub[n] = a
		}
	}
	var re []*regexp.Regexp
	for _, p := range c.Pinned {
		r, err := regexp.Compile(p)
		if err != nil {
			return c, err
		}
		re = append(re, r)
	}
	c.pinnedMatchers = re
	c.Applications = sub
	return c, nil
}
