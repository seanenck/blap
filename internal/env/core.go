// Package env handles environment settings
package env

import (
	"bytes"
	"errors"
	"runtime"
	"text/template"
)

// Values is the environment variables/values (for templating)
type Values[T any] struct {
	Name string
	Arch string
	OS   string
	Vars T
}

// NewValues create a new environment variable set
func NewValues[T any](name string, in T) (Values[T], error) {
	if name == "" {
		return Values[T]{}, errors.New("name must be set")
	}
	res := Values[T]{}
	res.Name = name
	res.Arch = runtime.GOARCH
	res.OS = runtime.GOOS
	res.Vars = in
	return res, nil
}

// Template will template a string
func (v Values[T]) Template(in string) (string, error) {
	t, err := template.New("t").Parse(in)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, v); err != nil {
		return "", err
	}
	return buf.String(), nil
}
