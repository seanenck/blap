// Package core handles templating value
package core

import (
	"bytes"
	"errors"
	"regexp"
	"runtime"
	"text/template"
)

var (
	templateRegexp = regexp.MustCompile(`{{(.*?)}}`)
	// BaseTemplate are common templating needs
	BaseTemplate = baseValues{
		runtime.GOARCH,
		runtime.GOOS,
	}
)

type (
	baseValues struct {
		Arch string
		OS   string
	}

	// Values is the environment variables/values (for templating)
	Values[T any] struct {
		baseValues
		Name string
		Vars T
	}
)

// NewValues create a new environment variable set
func NewValues[T any](name string, in T) (Values[T], error) {
	if name == "" {
		return Values[T]{}, errors.New("name must be set")
	}
	res := Values[T]{}
	res.baseValues = BaseTemplate
	res.Name = name
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
