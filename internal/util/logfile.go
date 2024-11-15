package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// AppendToLog handles simple log writing
func AppendToLog(logFile, msg string, parts ...any) error {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		m := fmt.Sprintf(msg, parts...)
		m = strings.TrimSpace(m)
		m = fmt.Sprintf("%s - %s", time.Now().Format("2006-01-02T15:04:05"), m)
		if _, err := fmt.Fprint(f, m); err != nil {
			return err
		}
	}
	return nil
}

// RotateLog handles a very simple log rotation
func RotateLog(logFile string, inSize int64, callback func()) error {
	if logFile != "" && PathExists(logFile) {
		info, err := os.Stat(logFile)
		if err != nil {
			return err
		}
		size := inSize
		switch {
		case size == 0:
			size = 10
		case size > 0:
		case size < 0:
			return fmt.Errorf("invalid log roll size, < 0 (have: %d)", size)
		}
		if info.Size() > size*1024*1024 {
			callback()
			old := fmt.Sprintf("%s.old", logFile)
			r, err := os.ReadFile(logFile)
			if err != nil {
				return err
			}
			if err := os.WriteFile(old, r, 0o644); err != nil {
				return err
			}
			if err := os.Remove(logFile); err != nil {
				return err
			}
		}
	}
	return nil
}
