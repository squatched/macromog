package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/aliases"
)

func TestResolveCharDir_Explicit(t *testing.T) {
	dir := t.TempDir()
	got, err := resolveCharDir(dir, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	abs, _ := filepath.Abs(dir)
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}

func TestResolveCharDir_ExplicitBadPath(t *testing.T) {
	_, err := resolveCharDir("/nonexistent/char", "", "")
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}
}

func TestResolveCharDir_SingleChar(t *testing.T) {
	ffxiDir, userDir, charDir := makeFFXITree(t, "a1b2c3d4")

	got, err := resolveCharDir("", "", ffxiDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = userDir
	abs, _ := filepath.Abs(charDir)
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}

func TestResolveCharDir_MultipleChars_NonTTY(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")

	_, err := resolveCharDir("", "", ffxiDir)
	if err == nil {
		t.Fatal("expected error for multiple chars on non-TTY stdin, got nil")
	}
}

func TestResolveCharDir_NoChars(t *testing.T) {
	ffxiDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(ffxiDir, "USER"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := resolveCharDir("", "", ffxiDir)
	if err == nil {
		t.Fatal("expected error for empty USER dir, got nil")
	}
}

func TestResolveCharDir_BadFFXIPath(t *testing.T) {
	_, err := resolveCharDir("", "", "/nonexistent/ffxi")
	if err == nil {
		t.Fatal("expected error for nonexistent ffxi path, got nil")
	}
}

func TestResolveCharDirs_Explicit(t *testing.T) {
	dir := t.TempDir()
	dirs, err := resolveCharDirs(dir, "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("got %d dirs, want 1", len(dirs))
	}
	abs, _ := filepath.Abs(dir)
	if dirs[0] != abs {
		t.Errorf("got %q, want %q", dirs[0], abs)
	}
}

func TestResolveCharDirs_CharAndAllMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	_, err := resolveCharDirs(dir, "", "", true)
	if err == nil {
		t.Error("expected error when --char-dir and --all are both set, got nil")
	}
}

func TestResolveCharDirs_All(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")

	dirs, err := resolveCharDirs("", "", ffxiDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dirs) != 2 {
		t.Errorf("got %d dirs, want 2", len(dirs))
	}
}

func TestResolveCharDirs_AllEmpty(t *testing.T) {
	ffxiDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(ffxiDir, "USER"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := resolveCharDirs("", "", ffxiDir, true)
	if err == nil {
		t.Fatal("expected error for empty USER dir with --all, got nil")
	}
}

func TestResolveCharDir_ByName(t *testing.T) {
	ffxiDir, userDir, charDir := makeFFXITree(t, "a1b2c3d4")

	doc := aliases.Document{Version: 1, Chars: map[string]aliases.Entry{"a1b2c3d4": {Name: "Squatched"}}}
	if err := aliases.Save(userDir, doc); err != nil {
		t.Fatal(err)
	}

	got, err := resolveCharDir("", "Squatched", ffxiDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	abs, _ := filepath.Abs(charDir)
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}

func TestResolveCharDir_ByName_CaseInsensitive(t *testing.T) {
	ffxiDir, userDir, _ := makeFFXITree(t, "a1b2c3d4")

	doc := aliases.Document{Version: 1, Chars: map[string]aliases.Entry{"a1b2c3d4": {Name: "Squatched"}}}
	if err := aliases.Save(userDir, doc); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveCharDir("", "squatched", ffxiDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveCharDir_ByName_NotFound(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")

	if _, err := resolveCharDir("", "Nobody", ffxiDir); err == nil {
		t.Error("expected error for unknown name, got nil")
	}
}

func TestResolveCharDirs_CharDirAndCharNameMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	if _, err := resolveCharDirs(dir, "Squatched", "", false); err == nil {
		t.Error("expected error when --char-dir and --char-name are both set, got nil")
	}
}

func TestResolveCharDirs_CharNameAndAllMutuallyExclusive(t *testing.T) {
	ffxiDir, userDir, _ := makeFFXITree(t, "a1b2c3d4")

	doc := aliases.Document{Version: 1, Chars: map[string]aliases.Entry{"a1b2c3d4": {Name: "Squatched"}}}
	if err := aliases.Save(userDir, doc); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveCharDirs("", "Squatched", ffxiDir, true); err == nil {
		t.Error("expected error when --char-name and --all are both set, got nil")
	}
}

func TestResolveCharDir_ByName_FutureVersion(t *testing.T) {
	ffxiDir, userDir, charDir := makeFFXITree(t, "a1b2c3d4")

	// Write a future-version characters.yml that still has the alias we need.
	content := "version: 99\nchars:\n  a1b2c3d4:\n    name: Squatched\n    future_field: something\n"
	if err := os.WriteFile(filepath.Join(userDir, "characters.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should warn but still resolve successfully.
	got, err := resolveCharDir("", "Squatched", ffxiDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	abs, _ := filepath.Abs(charDir)
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}

func TestResolveCharDir_ByName_DirMissing(t *testing.T) {
	ffxiDir, userDir, _ := makeFFXITree(t, "a1b2c3d4")

	// Alias points to a real char ID, but then we delete the directory.
	doc := aliases.Document{Version: 1, Chars: map[string]aliases.Entry{"a1b2c3d4": {Name: "Squatched"}}}
	if err := aliases.Save(userDir, doc); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(userDir, "a1b2c3d4")); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveCharDir("", "Squatched", ffxiDir); err == nil {
		t.Error("expected error when aliased directory is missing, got nil")
	}
}

func TestParseSelection_Single(t *testing.T) {
	got, err := parseSelection("2", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{1}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_Multi(t *testing.T) {
	got, err := parseSelection("1,3", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 2}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_Range(t *testing.T) {
	got, err := parseSelection("1-3", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1, 2}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_All(t *testing.T) {
	got, err := parseSelection("all", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1, 2}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_Mixed(t *testing.T) {
	got, err := parseSelection("1,3-4", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 2, 3}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_Dedup(t *testing.T) {
	got, err := parseSelection("1,1,2", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_OutOfRange(t *testing.T) {
	_, err := parseSelection("5", 3)
	if err == nil {
		t.Error("expected error for out-of-range selection, got nil")
	}
}

func TestParseSelection_InvalidRange(t *testing.T) {
	_, err := parseSelection("3-1", 3)
	if err == nil {
		t.Error("expected error for reversed range, got nil")
	}
}

func TestParseSelection_NonNumeric(t *testing.T) {
	_, err := parseSelection("abc", 3)
	if err == nil {
		t.Error("expected error for non-numeric input, got nil")
	}
}

func TestParseSelection_EmptyString(t *testing.T) {
	got, err := parseSelection("", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("empty input: got %v, want []", got)
	}
}

func TestParseSelection_OnlyCommas(t *testing.T) {
	got, err := parseSelection(",,,", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("only commas: got %v, want []", got)
	}
}

func TestParseSelection_ExactlyMax(t *testing.T) {
	got, err := parseSelection("3", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{2}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseSelection_Zero(t *testing.T) {
	_, err := parseSelection("0", 3)
	if err == nil {
		t.Error("expected error for 0 (below min 1), got nil")
	}
}

func TestParseSelection_Whitespace(t *testing.T) {
	got, err := parseSelection(" 1 , 2 ", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{0, 1}
	if !equalInts(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// makeFFXITree creates a minimal fake FFXI install directory tree with the
// given character IDs and returns (ffxiDir, userDir, firstCharDir).
func makeFFXITree(t *testing.T, charIDs ...string) (ffxiDir, userDir, firstCharDir string) {
	t.Helper()
	ffxiDir = t.TempDir()
	userDir = filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i, id := range charIDs {
		dir := filepath.Join(userDir, id)
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			firstCharDir = dir
		}
	}
	return ffxiDir, userDir, firstCharDir
}
