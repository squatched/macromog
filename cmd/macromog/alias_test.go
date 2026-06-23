package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/aliases"
)

// makeAliasUserDir creates a USER directory with one valid character folder.
func makeAliasUserDir(t *testing.T, charID string) string {
	t.Helper()
	userDir := t.TempDir()
	charDir := filepath.Join(userDir, charID)
	if err := os.Mkdir(charDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	return userDir
}

func TestRunAlias_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		if got := runAlias([]string{flag}, newTextPrinter()); got != 0 {
			t.Errorf("runAlias(%s) = %d, want 0", flag, got)
		}
	}
}

func TestRunAlias_NoArgs(t *testing.T) {
	if got := runAlias(nil, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(nil) = %d, want 1", got)
	}
}

func TestRunAlias_BadFFXIPath(t *testing.T) {
	args := []string{"--ffxi-path", "/nonexistent", "abc123", "Squatched"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(bad ffxi-path) = %d, want 1", got)
	}
}

func TestRunAlias_Set_InvalidCharID(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	args := []string{"--ffxi-path", ffxiDir, "notachar", "Squatched"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(invalid char id) = %d, want 1", got)
	}
}

func TestRunAlias_Set_Success(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charDir := filepath.Join(userDir, "abc123")
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "abc123", "Squatched"}
	if got := runAlias(args, newTextPrinter()); got != 0 {
		t.Fatalf("runAlias(set) = %d, want 0", got)
	}

	doc, err := aliases.Load(userDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if entry := doc.Chars["abc123"]; entry.Name != "Squatched" {
		t.Errorf("alias = %q, want %q", entry.Name, "Squatched")
	}
}

func TestRunAlias_Set_Overwrites(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charDir := filepath.Join(userDir, "abc123")
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"OldName", "Squatched"} {
		args := []string{"--ffxi-path", ffxiDir, "abc123", name}
		if got := runAlias(args, newTextPrinter()); got != 0 {
			t.Fatalf("runAlias(set %s) = %d, want 0", name, got)
		}
	}

	doc, _ := aliases.Load(userDir)
	if entry := doc.Chars["abc123"]; entry.Name != "Squatched" {
		t.Errorf("alias after overwrite = %q, want %q", entry.Name, "Squatched")
	}
}

func TestRunAlias_Remove_Success(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	doc := aliases.Document{Version: 1, Chars: map[string]aliases.Entry{"abc123": {Name: "Squatched"}}}
	if err := aliases.Save(userDir, doc); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "--remove", "abc123"}
	if got := runAlias(args, newTextPrinter()); got != 0 {
		t.Fatalf("runAlias(remove) = %d, want 0", got)
	}

	loaded, _ := aliases.Load(userDir)
	if _, ok := loaded.Chars["abc123"]; ok {
		t.Error("alias still present after remove")
	}
}

func TestRunAlias_Remove_NotFound(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "--remove", "abc123"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(remove missing) = %d, want 1", got)
	}
}

func TestRunAlias_Remove_NoArgs(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "--remove"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(--remove no arg) = %d, want 1", got)
	}
}

func TestRunAlias_Set_EmptyName(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charDir := filepath.Join(userDir, "abc123")
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "abc123", ""}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(empty name) = %d, want 1", got)
	}
}

func TestRunAlias_InvalidYAML(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charDir := filepath.Join(userDir, "abc123")
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userDir, "characters.yml"),
		[]byte(":\t not valid yaml at all"), 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "abc123", "Squatched"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(corrupt YAML) = %d, want 1", got)
	}
}

func TestRunAlias_JSON_Set(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charDir := filepath.Join(userDir, "abc123")
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	args := []string{"--ffxi-path", ffxiDir, "abc123", "Squatched"}
	if got := runAlias(args, p); got != 0 {
		t.Fatalf("runAlias(JSON set) = %d, want 0", got)
	}
	if !strings.Contains(buf.String(), `"char_id"`) {
		t.Errorf("JSON output missing char_id field:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "Squatched") {
		t.Errorf("JSON output missing name:\n%s", buf.String())
	}
}

func TestRunAlias_FutureVersion_BlocksSet(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charDir := filepath.Join(userDir, "abc123")
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userDir, "characters.yml"),
		[]byte("version: 99\nchars:\n  abc123:\n    name: Old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "abc123", "New"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(future version) = %d, want 1", got)
	}
}

func TestRunAlias_FutureVersion_BlocksRemove(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userDir, "characters.yml"),
		[]byte("version: 99\nchars:\n  abc123:\n    name: Old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	args := []string{"--ffxi-path", ffxiDir, "--remove", "abc123"}
	if got := runAlias(args, newTextPrinter()); got != 1 {
		t.Errorf("runAlias(remove future version) = %d, want 1", got)
	}
}
