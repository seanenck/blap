// Package util handles pathing
package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathExists indicates if a file exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// ResolveDirectory will resolve a directory from '~/' to $HOME
func ResolveDirectory(dir string) string {
	isHome := fmt.Sprintf("~%c", os.PathSeparator)
	if !strings.HasPrefix(dir, isHome) {
		return dir
	}
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return dir
	}
	return filepath.Join(h, strings.TrimPrefix(dir, isHome))
}
