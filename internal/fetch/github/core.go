// Package github has common definitions
package github

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type (
	// ErrorResponse is the error from github
	ErrorResponse struct {
		Message       string `json:"message"`
		Documentation string `json:"documentation_url"`
	}

	// WrapperError indicates a download error (from github specifically)
	WrapperError struct {
		Code   int
		Status string
		Body   []byte
		URL    string
	}
)

// Error is the interface definition for fetch errors
func (e *WrapperError) Error() string {
	components := make(map[string]string)
	for k, v := range map[string]string{
		"code":   fmt.Sprintf("%d", e.Code),
		"status": e.Status,
		"url":    e.URL,
	} {
		components[k] = v
	}
	if len(e.Body) > 0 {
		var resp ErrorResponse
		err := json.Unmarshal(e.Body, &resp)
		if err == nil {
			components["message"] = resp.Message
			components["doc"] = resp.Documentation
		} else {
			components["unmarshal"] = err.Error()
		}
	}

	var msg []string
	for k, v := range components {
		if strings.TrimSpace(v) != "" {
			msg = append(msg, fmt.Sprintf("%s: %s", k, v))
		}
	}
	sort.Strings(msg)
	return strings.Join(msg, "\n")
}
