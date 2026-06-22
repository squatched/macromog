package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_NoArgs(t *testing.T) {
	if got := run([]string{"macromog"}); got != 1 {
		t.Errorf("run(no args) = %d, want 1", got)
	}
}

func TestRun_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h", "help"} {
		if got := run([]string{"macromog", flag}); got != 0 {
			t.Errorf("run(%s) = %d, want 0", flag, got)
		}
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	if got := run([]string{"macromog", "bogus"}); got != 1 {
		t.Errorf("run(unknown) = %d, want 1", got)
	}
}

func TestRun_ExportMissingChar(t *testing.T) {
	if got := run([]string{"macromog", "export"}); got != 1 {
		t.Errorf("run(export) = %d, want 1", got)
	}
}

func TestRunValidate_NoArgs(t *testing.T) {
	if got := runValidate(nil); got != 1 {
		t.Errorf("runValidate(nil) = %d, want 1", got)
	}
}

func TestRunValidate_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		if got := runValidate([]string{flag}); got != 0 {
			t.Errorf("runValidate(%s) = %d, want 0", flag, got)
		}
	}
}

func TestRunValidate_FileNotFound(t *testing.T) {
	if got := runValidate([]string{"/nonexistent/path/file.yaml"}); got != 1 {
		t.Errorf("runValidate(missing file) = %d, want 1", got)
	}
}

func TestRunValidate_ValidFile(t *testing.T) {
	f := writeTemp(t, "version: 1\nbooks: {}\n")
	if got := runValidate([]string{f}); got != 0 {
		t.Errorf("runValidate(valid file) = %d, want 0", got)
	}
}

func TestRunValidate_InvalidFile(t *testing.T) {
	f := writeTemp(t, "version: 2\nbooks: {}\n")
	if got := runValidate([]string{f}); got != 1 {
		t.Errorf("runValidate(invalid file) = %d, want 1", got)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return filepath.Clean(f.Name())
}
