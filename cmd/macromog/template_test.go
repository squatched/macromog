package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTemplate_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		if got := runTemplate([]string{flag}, newTextPrinter()); got != 0 {
			t.Errorf("runTemplate(%s) = %d, want 0", flag, got)
		}
	}
}

func TestRunTemplate_NoArgs(t *testing.T) {
	// No output path → write template to stdout; command must succeed.
	if got := runTemplate(nil, newTextPrinter()); got != 0 {
		t.Errorf("runTemplate(nil) = %d, want 0", got)
	}
}

func TestRunTemplate_InvalidScope(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yml")
	if got := runTemplate([]string{"--scope", "B0", out}, newTextPrinter()); got != 1 {
		t.Errorf("runTemplate(bad scope) = %d, want 1", got)
	}
}

func TestRunTemplate_FullScope(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "full.yml")
	if got := runTemplate([]string{out}, newTextPrinter()); got != 0 {
		t.Errorf("runTemplate(full) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "level: full") {
		t.Errorf("full template missing scope level: %s", data)
	}
	if !strings.Contains(string(data), "books:") {
		t.Errorf("full template missing books: %s", data)
	}
}

func TestRunTemplate_BookScope(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "book.yml")
	if got := runTemplate([]string{"--scope", "B1", out}, newTextPrinter()); got != 0 {
		t.Errorf("runTemplate(book scope) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "level: book") {
		t.Errorf("expected book scope in output: %s", s)
	}
	// Only book 1 should be present; book 33 must not appear.
	if strings.Contains(s, "33:") {
		t.Errorf("scoped book template should not contain book 33: %s", s)
	}
	if !strings.Contains(s, "books:") {
		t.Errorf("book 1 template missing books section: %s", s)
	}
}

func TestRunTemplate_MacroScope(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "macro.yml")
	args := []string{"--scope", "B1S3A1,C2", out}
	if got := runTemplate(args, newTextPrinter()); got != 0 {
		t.Errorf("runTemplate(macro scope) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "level: macro") {
		t.Errorf("expected macro scope in output: %s", s)
	}
	if !strings.Contains(s, "alt:") || !strings.Contains(s, "ctrl:") {
		t.Errorf("expected both alt and ctrl in output: %s", s)
	}
}

func TestRunTemplate_CharName(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "char.yml")
	args := []string{"--char-name", "Kupomog", "--scope", "B1", out}
	if got := runTemplate(args, newTextPrinter()); got != 0 {
		t.Errorf("runTemplate(char-name) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Kupomog") {
		t.Errorf("character name not in output: %s", data)
	}
}

func TestRunTemplate_SetScope(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "set.yml")
	if got := runTemplate([]string{"--scope", "B1S3", out}, newTextPrinter()); got != 0 {
		t.Errorf("runTemplate(set scope) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "level: set") {
		t.Errorf("expected set scope in output: %s", s)
	}
	// Only book 1 / set 3 should be present.
	if !strings.Contains(s, "books:") {
		t.Errorf("template missing books section: %s", s)
	}
	if !strings.Contains(s, "sets:") {
		t.Errorf("template missing sets section: %s", s)
	}
	// Scope selections must reference book 1, set 3.
	if !strings.Contains(s, "book: 1") {
		t.Errorf("template missing book 1 selection: %s", s)
	}
	if !strings.Contains(s, "set: 3") {
		t.Errorf("template missing set 3 selection: %s", s)
	}
}

func TestRunTemplate_ScopeAfterOutputPath(t *testing.T) {
	// Flags placed after the positional output path must be parsed correctly.
	// Go's flag package stops at the first non-flag arg, so a second parse
	// pass is required to catch trailing flags.
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yml")
	args := []string{out, "--scope", "B1S3A1-5"}
	if got := runTemplate(args, newTextPrinter()); got != 0 {
		t.Fatalf("runTemplate(scope after path) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "level: macro") {
		t.Errorf("scope after positional: expected macro scope, got:\n%s", s)
	}
	if strings.Contains(s, "level: full") {
		t.Errorf("scope after positional: scope was ignored (level: full):\n%s", s)
	}
}

func TestRunTemplate_MultipleScopes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "multi.yml")
	args := []string{"--scope", "B1", "--scope", "B2", out}
	if got := runTemplate(args, newTextPrinter()); got != 0 {
		t.Fatalf("runTemplate(multi scope) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "book: 1") || !strings.Contains(s, "book: 2") {
		t.Errorf("multi-scope template missing expected books:\n%s", s)
	}
}

func TestRunTemplate_WriteFails(t *testing.T) {
	// Point output at a non-existent directory.
	out := "/nonexistent/dir/out.yml"
	if got := runTemplate([]string{out}, newTextPrinter()); got != 1 {
		t.Errorf("runTemplate(bad path) = %d, want 1", got)
	}
}
