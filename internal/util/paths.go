// Package util handles pathing
package util

import (
	"errors"
	"os"
)

// PathExists indicates if a file exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
