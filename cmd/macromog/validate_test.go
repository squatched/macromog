package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidate_MissingVersionField(t *testing.T) {
	f := writeTemp(t, "books: {}\n")
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(no version) = %d, want 1", got)
	}
}

func TestRunValidate_MissingBooksField(t *testing.T) {
	f := writeTemp(t, "version: 1\n")
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(no books) = %d, want 1", got)
	}
}

func TestRunValidate_JSON_Valid(t *testing.T) {
	f := writeTemp(t, "version: 1\nbooks: {}\n")
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runValidate([]string{f}, p); got != 0 {
		t.Fatalf("runValidate(JSON valid) = %d, want 0", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"valid": true`) {
		t.Errorf("JSON output missing valid:true:\n%s", s)
	}
	if !strings.Contains(s, `"file"`) {
		t.Errorf("JSON output missing file field:\n%s", s)
	}
}

func TestRunValidate_JSON_Invalid(t *testing.T) {
	f := writeTemp(t, "version: 99\nbooks: {}\n")
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runValidate([]string{f}, p); got != 1 {
		t.Fatalf("runValidate(JSON invalid) = %d, want 1", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"valid": false`) {
		t.Errorf("JSON output missing valid:false:\n%s", s)
	}
	if !strings.Contains(s, `"errors"`) {
		t.Errorf("JSON output missing errors field:\n%s", s)
	}
}

func TestRunValidate_JSON_ReportsFilePath(t *testing.T) {
	f := writeTemp(t, "version: 1\nbooks: {}\n")
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	runValidate([]string{f}, p)
	if !strings.Contains(buf.String(), filepath.Base(f)) {
		t.Errorf("JSON output does not contain filename %q:\n%s", f, buf.String())
	}
}

// Pathological: empty file — both version and books absent.
func TestRunValidate_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yml")
	_ = os.WriteFile(path, []byte{}, 0o644)
	if got := runValidate([]string{path}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(empty) = %d, want 1", got)
	}
}

// Pathological: binary bytes that cause a YAML parse error.
func TestRunValidate_BinaryGarbage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "garbage.yml")
	_ = os.WriteFile(path, []byte{0x00, 0xFF, 0xFE, 0x80, 0x01}, 0o644)
	if got := runValidate([]string{path}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(binary garbage) = %d, want 1", got)
	}
}

// Pathological: macro name exceeding 8 characters.
func TestRunValidate_MacroNameTooLong(t *testing.T) {
	yaml := "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          0:\n            name: TooLongNm\n            contents: []\n"
	f := writeTemp(t, yaml)
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(name too long) = %d, want 1", got)
	}
}

// Pathological: macro contents exceeding the 6-line limit.
func TestRunValidate_TooManyLines(t *testing.T) {
	yaml := "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          0:\n            name: Test\n            contents:\n              - /line1\n              - /line2\n              - /line3\n              - /line4\n              - /line5\n              - /line6\n              - /line7\n"
	f := writeTemp(t, yaml)
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(too many lines) = %d, want 1", got)
	}
}

// Pathological: a macro line exceeding 60 characters.
func TestRunValidate_MacroLineTooLong(t *testing.T) {
	line := strings.Repeat("x", 61)
	yaml := "version: 1\nbooks:\n  1:\n    sets:\n      1:\n        ctrl:\n          0:\n            name: Test\n            contents:\n              - " + line + "\n"
	f := writeTemp(t, yaml)
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(line too long) = %d, want 1", got)
	}
}

// Pathological: book index outside the allowed 1–40 range.
func TestRunValidate_BookIndexOutOfRange(t *testing.T) {
	f := writeTemp(t, "version: 1\nbooks:\n  41:\n    sets: {}\n")
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(book out of range) = %d, want 1", got)
	}
}

// Pathological: set index outside the allowed 1–10 range.
func TestRunValidate_SetIndexOutOfRange(t *testing.T) {
	f := writeTemp(t, "version: 1\nbooks:\n  1:\n    sets:\n      11:\n        ctrl: {}\n        alt: {}\n")
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(set out of range) = %d, want 1", got)
	}
}

// Pathological: exported_at value that is not a valid RFC3339 timestamp.
func TestRunValidate_BadTimestamp(t *testing.T) {
	f := writeTemp(t, "version: 1\nexported_at: not-a-timestamp\nbooks: {}\n")
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(bad timestamp) = %d, want 1", got)
	}
}

// Pathological: book name containing non-alphanumeric characters.
func TestRunValidate_BookNameInvalidChars(t *testing.T) {
	f := writeTemp(t, "version: 1\nbooks:\n  1:\n    name: \"invalid!\"\n    sets: {}\n")
	if got := runValidate([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runValidate(book name invalid chars) = %d, want 1", got)
	}
}
