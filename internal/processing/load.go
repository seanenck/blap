// Package processing loads configs
package processing

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/seanenck/blap/internal/cli"
	"github.com/seanenck/blap/internal/core"
	"github.com/seanenck/blap/internal/logging"
)

func doDecode[T any](in string, o T) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	decoder := toml.NewDecoder(bytes.NewReader(data))
	md, err := decoder.Decode(o)
	if err != nil {
		return err
	}
	unknown := md.Undecoded()
	if len(unknown) > 0 {
		return fmt.Errorf("unknown fields: %v", unknown)
	}
	return nil
}

// Load will initialize the configuration from a file
func Load(input string, context cli.Settings) (Configuration, error) {
	c := Configuration{}
	c.handler = &processHandler{}
	c.context = context
	c.Apps = make(map[string]core.Application)
	if err := doDecode(input, &c); err != nil {
		return c, err
	}
	c.logFile = c.Logging.File.String()
	c.dir = c.Directory.String()
	checkAddApp := func(name string, a core.Application) (bool, error) {
		if err := a.Flags.Check(); err != nil {
			return false, err
		}
		if a.Flags.Pin() {
			c.Pinned = append(c.Pinned, name)
		}
		return a.Enabled(), nil
	}
	logDebug := func(msg string, args ...any) {
		c.context.LogDebug(logging.ConfigCategory, msg, args...)
	}
	if len(c.Include) > 0 {
		var including []string
		for _, i := range c.Include {
			r := i.String()
			res := []string{r}
			logDebug("including: %s\n", i)
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
			logDebug("loading included: %s\n", include)
			type included struct {
				Apps   core.AppSet
				Flags  core.FlagSet
				Pinned core.Pinned
			}
			var apps included
			if err := doDecode(include, &apps); err != nil {
				return c, err
			}
			if err := apps.Flags.Check(); err != nil {
				return Configuration{}, err
			}
			if apps.Flags.Skipped() {
				if apps.Flags.Pin() {
					for k := range apps.Apps {
						c.Pinned = append(c.Pinned, k)
					}
				}
				continue
			}
			c.Pinned = append(c.Pinned, apps.Pinned...)
			for k, v := range apps.Apps {
				ok, err := checkAddApp(k, v)
				if err != nil {
					return Configuration{}, err
				}
				if !ok {
					continue
				}
				if _, ok := c.Apps[k]; ok {
					return c, fmt.Errorf("%s is overwritten by config: %s", k, include)
				}
				c.Apps[k] = v
			}
		}
	}
	canFilter := context.FilterApplications()
	sub := make(map[string]core.Application)
	for n, a := range c.Apps {
		ok, err := checkAddApp(n, a)
		if err != nil {
			return Configuration{}, err
		}
		if !ok {
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
	var knownPins []string
	for _, p := range c.Pinned {
		if slices.Contains(knownPins, p) {
			continue
		}
		r, err := regexp.Compile(p)
		if err != nil {
			return c, err
		}
		re = append(re, r)
		knownPins = append(knownPins, p)
	}
	c.Pinned = knownPins
	c.pinnedMatchers = re
	c.Apps = sub
	return c, nil
}

// List will simply list information
func (c Configuration) List(w io.Writer) error {
	if c.Apps != nil {
		var keys []string
		for k := range c.Apps {
			keys = append(keys, k)
		}
		if err := list(w, "app", keys); err != nil {
			return err
		}
	}
	return list(w, "pin", c.Pinned)
}

func list(w io.Writer, header string, keys []string) error {
	sort.Strings(keys)
	for _, item := range keys {
		if _, err := fmt.Fprintf(w, "(%s) -> %s\n", header, item); err != nil {
			return err
		}
	}
	return nil
}
