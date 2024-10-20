// Package util handles pathing
package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const alpha = "abcdefghijklmnopqrstuvwxyz"

var allAllowedChars = fmt.Sprintf("%s%s%s%s", alpha, strings.ToUpper(alpha), "0123456789", ".-_")

// PathExists indicates if a file exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// CleanFileName will strip unwanted characters from a file system name
func CleanFileName(file string) string {
	res := ""
	for _, r := range file {
		if strings.Contains(allAllowedChars, string(r)) {
			res = fmt.Sprintf("%s%c", res, r)
		}
	}
	return res
}
