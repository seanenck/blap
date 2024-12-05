package core

import (
	"fmt"
	"strings"
)

// Version is a tag that could be a semantic version
type Version string

// Major is the major component
func (v Version) Major() string {
	m, _, _, _ := v.parse()
	return m
}

// Minor is the minor component
func (v Version) Minor() string {
	_, m, _, _ := v.parse()
	return m
}

// Patch is the patch component
func (v Version) Patch() string {
	_, _, p, _ := v.parse()
	return p
}

// Remainder is anything after patch
func (v Version) Remainder() string {
	_, _, _, r := v.parse()
	return r
}

// Full is the major.minor.patch.remainder (- prefix 'v')
func (v Version) Full() string {
	major, minor, patch, remainder := v.parse()
	if major == "" {
		return ""
	}
	if minor == "" {
		return major
	}
	if patch == "" {
		return fmt.Sprintf("%s.%s", major, minor)
	}
	if remainder == "" {
		return fmt.Sprintf("%s.%s.%s", major, minor, patch)
	}
	return strings.Join([]string{major, minor, patch, remainder}, ".")
}

func (v Version) parse() (string, string, string, string) {
	parts := strings.Split(string(v), ".")
	major := strings.TrimPrefix(parts[0], "v")
	var minor, patch, left string
	if len(parts) > 1 {
		minor = parts[1]
		if len(parts) > 2 {
			patch = parts[2]
			if len(parts) > 3 {
				left = strings.Join(parts[3:], ".")
			}
		}
	}
	return major, minor, patch, left
}
