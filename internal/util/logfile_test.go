package util_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/blap/internal/util"
)

func setupTeardown() func() {
	os.RemoveAll("testdata")
	os.Mkdir("testdata", 0o755)
	return func() {
		os.RemoveAll("testdata")
	}
}

func TestAppendTo(t *testing.T) {
	defer setupTeardown()()
	if err := util.AppendToLog("", "ab"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	files, _ := os.ReadDir("testdata")
	if len(files) != 0 {
		t.Errorf("invalid dir: %v", files)
	}
	log := filepath.Join("testdata", "log")
	if err := util.AppendToLog(log, "ab"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	files, _ = os.ReadDir("testdata")
	if len(files) != 1 {
		t.Errorf("invalid dir: %v", files)
	}
	logTwo := filepath.Join("testdata", "log2")
	if err := util.AppendToLog(logTwo, "ab"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := util.AppendToLog(logTwo, "xy"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	files, _ = os.ReadDir("testdata")
	if len(files) != 2 {
		t.Errorf("invalid dir: %v", files)
	}
	d, _ := os.ReadFile(log)
	str := string(d)
	if !strings.Contains(str, "ab") || strings.Contains(str, "xy") {
		t.Errorf("invalid data: %s", str)
	}
	d, _ = os.ReadFile(logTwo)
	str = string(d)
	if !strings.Contains(str, "ab") || !strings.Contains(str, "xy") {
		t.Errorf("invalid data: %s", str)
	}
}

func TestRotate(t *testing.T) {
	defer setupTeardown()()
	if err := util.RotateLog("", -1, func() {}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	log := filepath.Join("testdata", "log")
	if err := util.RotateLog(log, -1, func() {}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.WriteFile(log, []byte{}, 0o644)
	if err := util.RotateLog(log, -1, func() {}); err == nil || err.Error() != "invalid log roll size, < 0 (have: -1)" {
		t.Errorf("invalid error: %v", err)
	}
	rotated := false
	rotate := func() {
		rotated = true
	}
	if err := util.RotateLog(log, 0, rotate); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if rotated {
		t.Error("should NOT have rotated")
	}
	files, _ := os.ReadDir("testdata")
	if len(files) != 1 || files[0].Name() != "log" {
		t.Errorf("invalid dir: %v", files)
	}
	var buf []byte
	i := 0
	for i < (1*1024*1024 + 1024) {
		buf = append(buf, 1)
		i++
	}
	os.WriteFile(log, buf, 0o644)
	if err := util.RotateLog(log, 1, rotate); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !rotated {
		t.Error("should have rotated")
	}
	files, _ = os.ReadDir("testdata")
	if len(files) != 1 || files[0].Name() != "log.old" {
		t.Errorf("invalid dir: %v", files)
	}
}
