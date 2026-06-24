package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
)

func TestRunImport_Help(t *testing.T) {
	if got := runImport([]string{"--help"}, newTextPrinter()); got != 0 {
		t.Errorf("runImport(--help) = %d, want 0", got)
	}
}

func TestRunImport_NoArgs(t *testing.T) {
	if got := runImport(nil, newTextPrinter()); got != 1 {
		t.Errorf("runImport(nil) = %d, want 1", got)
	}
}

func TestRunImport_MissingFile(t *testing.T) {
	dir := t.TempDir()
	args := []string{"/nonexistent/macros.yml", dir}
	if got := runImport(args, newTextPrinter()); got != 1 {
		t.Errorf("runImport(missing file) = %d, want 1", got)
	}
}

func TestRunImport_MissingCharDir(t *testing.T) {
	dir := t.TempDir()
	f := writeImportTemp(t, dir, "v.yml", "version: 1\nbooks: {}\n")
	if got := runImport([]string{f}, newTextPrinter()); got != 1 {
		t.Errorf("runImport(no char dir) = %d, want 1", got)
	}
}

func TestRunImport_BadCharDir(t *testing.T) {
	dir := t.TempDir()
	f := writeImportTemp(t, dir, "v.yml", "version: 1\nbooks: {}\n")
	if got := runImport([]string{f, "/nonexistent/char"}, newTextPrinter()); got != 1 {
		t.Errorf("runImport(bad char dir) = %d, want 1", got)
	}
}

func TestRunImport_ValidationFails(t *testing.T) {
	dir := t.TempDir()
	f := writeImportTemp(t, dir, "bad.yml", "version: 99\nbooks: {}\n")
	if got := runImport([]string{f, dir}, newTextPrinter()); got != 1 {
		t.Errorf("runImport(invalid YAML) = %d, want 1", got)
	}
}

func TestRunImport_DryRun(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	doc, _ := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	data, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	destDir := t.TempDir()
	if got := runImport([]string{"--dry-run", "--no-backup", yamlPath, destDir}, newTextPrinter()); got != 0 {
		t.Errorf("runImport(--dry-run) = %d, want 0", got)
	}
	entries, _ := os.ReadDir(destDir)
	if len(entries) != 0 {
		t.Errorf("dry-run wrote %d file(s), want 0", len(entries))
	}
}

func TestRunImport_Success(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	doc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	destDir := t.TempDir()
	args := []string{"--no-backup", yamlPath, destDir}
	if got := runImport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runImport = %d, want 0", got)
	}

	// Re-export and verify expected content survived the round-trip
	destDoc, err := export.FromCharacterDir(export.Options{CharacterDir: destDir, Character: "char"})
	if err != nil {
		t.Fatalf("re-export: %v", err)
	}
	destData, _ := export.MarshalYAML(destDoc)
	if !strings.Contains(string(destData), "B33S1") {
		t.Errorf("missing B33S1 in re-exported YAML:\n%s", destData)
	}
}

func TestRunImport_CharFlag(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	doc, _ := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	data, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	destDir := t.TempDir()
	args := []string{"--no-backup", "--char-dir", destDir, yamlPath}
	if got := runImport(args, newTextPrinter()); got != 0 {
		t.Errorf("runImport(--char-dir flag) = %d, want 0", got)
	}
}

func TestRunImport_AllFlag(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")

	// Build a valid YAML from real testdata.
	src := testdata.CharDir()
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, _ := export.MarshalYAML(doc)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	// Seed each char dir with .dat files so backup/import has something to work with.
	for _, id := range []string{"a1b2c3d4", "e5f6a7b8"} {
		charDir := filepath.Join(ffxiDir, "USER", id)
		entries, _ := os.ReadDir(src)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			srcData, _ := os.ReadFile(filepath.Join(src, e.Name()))
			_ = os.WriteFile(filepath.Join(charDir, e.Name()), srcData, 0o644)
		}
	}

	args := []string{"--ffxi-path", ffxiDir, "--all", "--no-backup", yamlPath}
	if got := runImport(args, newTextPrinter()); got != 0 {
		t.Errorf("runImport(--all) = %d, want 0", got)
	}
}

func TestRunImport_WithScopeFlag(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	doc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	destDir := t.TempDir()
	// B33 is a subset of the full-scope YAML → no confirmation needed.
	args := []string{"--no-backup", "--scope", "B33", yamlPath, destDir}
	if got := runImport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runImport(--scope B33) = %d, want 0", got)
	}
}

func TestRunImport_InvalidScopeFlag(t *testing.T) {
	if got := runImport([]string{"--scope", "B41", "fake.yml", t.TempDir()}, newTextPrinter()); got != 1 {
		t.Errorf("runImport(bad scope) = %d, want 1", got)
	}
}

func TestConfirmScopeOverride_NoExceed(t *testing.T) {
	// Full-scope YAML; import scope B1 is narrower → no confirmation.
	yamlContent := "version: 1\nscope:\n  level: full\nbooks: {}\n"
	tmp := t.TempDir()
	path := filepath.Join(tmp, "m.yml")
	_ = os.WriteFile(path, []byte(yamlContent), 0o644)

	sc := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}}}
	confirmed, err := confirmScopeOverride(path, sc, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected confirmed=true for subset scope")
	}
}

func TestConfirmScopeOverride_ExceedsAndAccepts(t *testing.T) {
	// Book-scope B1 YAML; import scope B1,5 exceeds → confirmation needed.
	yamlContent := "version: 1\nscope:\n  level: book\n  selections:\n    - {book: 1}\nbooks: {}\n"
	tmp := t.TempDir()
	path := filepath.Join(tmp, "m.yml")
	_ = os.WriteFile(path, []byte(yamlContent), 0o644)

	sc := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}, {Book: 5}}}
	confirmed, err := confirmScopeOverride(path, sc, strings.NewReader("y\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected confirmed=true when user types 'y'")
	}
}

func TestConfirmScopeOverride_ExceedsAndRejects(t *testing.T) {
	yamlContent := "version: 1\nscope:\n  level: book\n  selections:\n    - {book: 1}\nbooks: {}\n"
	tmp := t.TempDir()
	path := filepath.Join(tmp, "m.yml")
	_ = os.WriteFile(path, []byte(yamlContent), 0o644)

	sc := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}, {Book: 5}}}
	confirmed, err := confirmScopeOverride(path, sc, strings.NewReader("n\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected confirmed=false when user types 'n'")
	}
}

func TestConfirmScopeOverride_ExceedsNoInput(t *testing.T) {
	yamlContent := "version: 1\nscope:\n  level: book\n  selections:\n    - {book: 1}\nbooks: {}\n"
	tmp := t.TempDir()
	path := filepath.Join(tmp, "m.yml")
	_ = os.WriteFile(path, []byte(yamlContent), 0o644)

	sc := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}, {Book: 5}}}
	confirmed, err := confirmScopeOverride(path, sc, strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected confirmed=false on EOF")
	}
}

func TestConfirmScopeOverride_FileNotFound(t *testing.T) {
	sc := scope.Scope{Level: scope.LevelFull}
	_, err := confirmScopeOverride("/nonexistent/file.yml", sc, nil)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunImport_JSON_Success(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	doc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	destDir := t.TempDir()
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runImport([]string{"--no-backup", yamlPath, destDir}, p); got != 0 {
		t.Fatalf("runImport(JSON) = %d, want 0", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"ok"`) {
		t.Errorf("JSON output missing ok field:\n%s", s)
	}
	if !strings.Contains(s, `"sets"`) {
		t.Errorf("JSON output missing sets field:\n%s", s)
	}
}

func TestRunImport_JSON_DryRun(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	doc, _ := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	data, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	destDir := t.TempDir()
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runImport([]string{"--dry-run", "--no-backup", yamlPath, destDir}, p); got != 0 {
		t.Fatalf("runImport(JSON dry-run) = %d, want 0", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"dry_run"`) {
		t.Errorf("JSON output missing dry_run field:\n%s", s)
	}
	if !strings.Contains(s, `"dry_run_sets"`) {
		t.Errorf("JSON output missing dry_run_sets field:\n%s", s)
	}
}

func TestRunImport_DryRun_All(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")

	src := testdata.CharDir()
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, _ := export.MarshalYAML(doc)
	tmp := t.TempDir()
	yamlPath := filepath.Join(tmp, "macros.yml")
	_ = os.WriteFile(yamlPath, data, 0o644)

	args := []string{"--ffxi-path", ffxiDir, "--all", "--dry-run", "--no-backup", yamlPath}
	if got := runImport(args, newTextPrinter()); got != 0 {
		t.Errorf("runImport(--dry-run --all) = %d, want 0", got)
	}
	// Neither char dir should have any new .dat files written.
	for _, id := range []string{"a1b2c3d4", "e5f6a7b8"} {
		charDir := filepath.Join(ffxiDir, "USER", id)
		entries, _ := os.ReadDir(charDir)
		var datCount int
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".dat") {
				datCount++
			}
		}
		// makeFFXITree writes exactly one mcr.dat per char; dry-run must not add more.
		if datCount > 1 {
			t.Errorf("dry-run wrote .dat files in %s (found %d)", id, datCount)
		}
	}
}

func TestConfirmScopeOverride_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.yml")
	_ = os.WriteFile(path, []byte(":\tnot valid yaml at all"), 0o644)

	sc := scope.Scope{Level: scope.LevelFull}
	_, err := confirmScopeOverride(path, sc, nil)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func writeImportTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
