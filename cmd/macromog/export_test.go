package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestRunExport_MissingChar(t *testing.T) {
	if got := runExport(nil); got != 1 {
		t.Errorf("runExport(nil) = %d, want 1", got)
	}
}

func TestRunExport_Help(t *testing.T) {
	if got := runExport([]string{"--help"}); got != 0 {
		t.Errorf("runExport(--help) = %d, want 0", got)
	}
}

func TestRunExport_BadCharDir(t *testing.T) {
	if got := runExport([]string{"--char", "/nonexistent/char"}); got != 1 {
		t.Errorf("runExport(bad char) = %d, want 1", got)
	}
}

func TestRunExport_Book33(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "book33.yml")
	args := []string{"--char", testdata.CharDir(), "-o", out}
	if got := runExport(args); got != 0 {
		t.Errorf("runExport = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "B33S1") {
		t.Errorf("missing B33S1 in output: %s", data)
	}
}

func TestRunExport_PositionalCharDir(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "book33.yml")
	if got := runExport([]string{testdata.CharDir(), out}); got != 0 {
		t.Errorf("runExport(positional) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "B33S1") {
		t.Errorf("missing B33S1 in output: %s", data)
	}
}
