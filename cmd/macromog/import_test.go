package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
	"github.com/squatched/macromog/internal/export"
)

func TestRunImport_Help(t *testing.T) {
	if got := runImport([]string{"--help"}); got != 0 {
		t.Errorf("runImport(--help) = %d, want 0", got)
	}
}

func TestRunImport_NoArgs(t *testing.T) {
	if got := runImport(nil); got != 1 {
		t.Errorf("runImport(nil) = %d, want 1", got)
	}
}

func TestRunImport_MissingFile(t *testing.T) {
	dir := t.TempDir()
	args := []string{"/nonexistent/macros.yml", dir}
	if got := runImport(args); got != 1 {
		t.Errorf("runImport(missing file) = %d, want 1", got)
	}
}

func TestRunImport_MissingCharDir(t *testing.T) {
	dir := t.TempDir()
	f := writeImportTemp(t, dir, "v.yml", "version: 1\nbooks: {}\n")
	if got := runImport([]string{f}); got != 1 {
		t.Errorf("runImport(no char dir) = %d, want 1", got)
	}
}

func TestRunImport_BadCharDir(t *testing.T) {
	dir := t.TempDir()
	f := writeImportTemp(t, dir, "v.yml", "version: 1\nbooks: {}\n")
	if got := runImport([]string{f, "/nonexistent/char"}); got != 1 {
		t.Errorf("runImport(bad char dir) = %d, want 1", got)
	}
}

func TestRunImport_ValidationFails(t *testing.T) {
	dir := t.TempDir()
	f := writeImportTemp(t, dir, "bad.yml", "version: 99\nbooks: {}\n")
	if got := runImport([]string{f, dir}); got != 1 {
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
	if got := runImport([]string{"--dry-run", "--no-backup", yamlPath, destDir}); got != 0 {
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
	if got := runImport(args); got != 0 {
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
	args := []string{"--no-backup", "--char", destDir, yamlPath}
	if got := runImport(args); got != 0 {
		t.Errorf("runImport(--char) = %d, want 0", got)
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
	if got := runImport(args); got != 0 {
		t.Errorf("runImport(--all) = %d, want 0", got)
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
