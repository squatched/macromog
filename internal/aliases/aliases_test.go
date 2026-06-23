package aliases

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileNotFound(t *testing.T) {
	doc, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Version != 1 {
		t.Errorf("version = %d, want 1", doc.Version)
	}
	if len(doc.Chars) != 0 {
		t.Errorf("chars = %v, want empty", doc.Chars)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := "version: 1\nchars:\n  abc123:\n    name: Squatched\n"
	if err := os.WriteFile(filepath.Join(dir, "characters.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Version != 1 {
		t.Errorf("version = %d, want 1", doc.Version)
	}
	if entry, ok := doc.Chars["abc123"]; !ok || entry.Name != "Squatched" {
		t.Errorf("chars[abc123] = %v, want {Squatched}", doc.Chars["abc123"])
	}
}

func TestLoad_FutureVersion_ReturnsDocAndError(t *testing.T) {
	dir := t.TempDir()
	// A future version may add new Entry fields; the name we understand is still there.
	content := "version: 2\nchars:\n  abc123:\n    name: Squatched\n    new_field: somevalue\n"
	if err := os.WriteFile(filepath.Join(dir, "characters.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := Load(dir)
	if err == nil {
		t.Fatal("expected FutureVersionError, got nil")
	}
	if !IsFutureVersion(err) {
		t.Errorf("expected FutureVersionError, got %T: %v", err, err)
	}
	// Document is still usable for read-only operations.
	if got := LookupName(doc, "abc123"); got != "Squatched" {
		t.Errorf("LookupName after future-version load = %q, want %q", got, "Squatched")
	}
}

func TestLoad_FutureVersion_BlocksWrite(t *testing.T) {
	dir := t.TempDir()
	content := "version: 2\nchars:\n  abc123:\n    name: Squatched\n"
	if err := os.WriteFile(filepath.Join(dir, "characters.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if !IsFutureVersion(err) {
		t.Fatalf("expected FutureVersionError, got %v", err)
	}
	// Callers that receive IsFutureVersion should not call Save.
	// This test documents the intended caller contract: check before writing.
}

func TestIsFutureVersion_OtherErrors(t *testing.T) {
	if IsFutureVersion(fmt.Errorf("some other error")) {
		t.Error("IsFutureVersion returned true for a non-FutureVersionError")
	}
	if IsFutureVersion(nil) {
		t.Error("IsFutureVersion returned true for nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "characters.yml"), []byte(":\t bad yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	doc := Document{
		Version: 1,
		Chars: map[string]Entry{
			"abc123": {Name: "Squatched"},
		},
	}
	if err := Save(dir, doc); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if entry, ok := loaded.Chars["abc123"]; !ok || entry.Name != "Squatched" {
		t.Errorf("round-trip: chars[abc123] = %v, want {Squatched}", loaded.Chars["abc123"])
	}
}

func TestResolve_Found(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	id, err := Resolve(doc, "Squatched")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "abc123" {
		t.Errorf("id = %q, want %q", id, "abc123")
	}
}

func TestResolve_CaseInsensitive(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	id, err := Resolve(doc, "squatched")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "abc123" {
		t.Errorf("id = %q, want %q", id, "abc123")
	}
}

func TestResolve_NotFound(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if _, err := Resolve(doc, "Nobody"); err == nil {
		t.Error("expected error for unknown name, got nil")
	}
}

func TestResolve_MultipleMatches(t *testing.T) {
	doc := Document{Chars: map[string]Entry{
		"abc123": {Name: "Squatched"},
		"def456": {Name: "squatched"},
	}}
	if _, err := Resolve(doc, "Squatched"); err == nil {
		t.Error("expected error for duplicate name, got nil")
	}
}

func TestResolve_EmptyDoc(t *testing.T) {
	doc := Document{}
	if _, err := Resolve(doc, "Squatched"); err == nil {
		t.Error("expected error for empty doc, got nil")
	}
}

func TestLookupName_Found(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if got := LookupName(doc, "abc123"); got != "Squatched" {
		t.Errorf("got %q, want %q", got, "Squatched")
	}
}

func TestLookupName_NotFound(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if got := LookupName(doc, "ffffff"); got != "ffffff" {
		t.Errorf("got %q, want %q", got, "ffffff")
	}
}

func TestLookupName_EmptyName(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: ""}}}
	if got := LookupName(doc, "abc123"); got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestLookupName_NilChars(t *testing.T) {
	doc := Document{Version: 1}
	if got := LookupName(doc, "abc123"); got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "characters.yml"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error loading empty file: %v", err)
	}
	if doc.Version != 0 {
		t.Errorf("version = %d, want 0 (zero value from empty file)", doc.Version)
	}
}

func TestSave_NonexistentDir(t *testing.T) {
	doc := Document{Version: 1, Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if err := Save("/nonexistent/path", doc); err == nil {
		t.Error("expected error saving to nonexistent dir, got nil")
	}
}

func TestResolve_EmptyName(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if _, err := Resolve(doc, ""); err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

func TestResolve_WhitespaceName(t *testing.T) {
	doc := Document{Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if _, err := Resolve(doc, "   "); err == nil {
		t.Error("expected error for whitespace-only name, got nil")
	}
}

// TestAliasesScoped verifies that aliases in one USER directory are independent
// of aliases in another — the multi-install case.
func TestAliasesScoped(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	doc1 := Document{Version: 1, Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	doc2 := Document{Version: 1, Chars: map[string]Entry{"abc123": {Name: "AltChar"}}}

	if err := Save(dir1, doc1); err != nil {
		t.Fatalf("Save dir1: %v", err)
	}
	if err := Save(dir2, doc2); err != nil {
		t.Fatalf("Save dir2: %v", err)
	}

	loaded1, err := Load(dir1)
	if err != nil {
		t.Fatalf("Load dir1: %v", err)
	}
	loaded2, err := Load(dir2)
	if err != nil {
		t.Fatalf("Load dir2: %v", err)
	}

	if got := LookupName(loaded1, "abc123"); got != "Squatched" {
		t.Errorf("dir1 alias = %q, want %q", got, "Squatched")
	}
	if got := LookupName(loaded2, "abc123"); got != "AltChar" {
		t.Errorf("dir2 alias = %q, want %q", got, "AltChar")
	}
}

// TestSaveScoped verifies that saving to one USER dir does not create or modify
// the characters.yml in a different USER dir.
func TestSaveScoped(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	doc := Document{Version: 1, Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if err := Save(dir1, doc); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// dir2 should have no characters.yml at all.
	if _, err := os.Stat(filepath.Join(dir2, "characters.yml")); !os.IsNotExist(err) {
		t.Errorf("expected no characters.yml in dir2, got err=%v", err)
	}
}

// TestResolveScoped verifies that Resolve against one install's doc does not
// find aliases defined only in a different install's doc.
func TestResolveScoped(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	doc1 := Document{Version: 1, Chars: map[string]Entry{"abc123": {Name: "Squatched"}}}
	if err := Save(dir1, doc1); err != nil {
		t.Fatalf("Save dir1: %v", err)
	}

	loaded2, err := Load(dir2) // dir2 has no characters.yml
	if err != nil {
		t.Fatalf("Load dir2: %v", err)
	}

	if _, err := Resolve(loaded2, "Squatched"); err == nil {
		t.Error("expected error resolving name from a different install's USER dir, got nil")
	}
}
