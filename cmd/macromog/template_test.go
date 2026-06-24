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
	if got := runTemplate(nil, newTextPrinter()); got != 1 {
		t.Errorf("runTemplate(nil) = %d, want 1", got)
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

func TestRunTemplate_WriteFails(t *testing.T) {
	// Point output at a non-existent directory.
	out := "/nonexistent/dir/out.yml"
	if got := runTemplate([]string{out}, newTextPrinter()); got != 1 {
		t.Errorf("runTemplate(bad path) = %d, want 1", got)
	}
}
