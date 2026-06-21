package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
	"github.com/squatched/macromog/internal/export"
)

func TestImport_ValidationFailure(t *testing.T) {
	dir := t.TempDir()
	f := writeYAML(t, dir, "bad.yml", "version: 99\nbooks: {}\n")
	_, err := Import(Options{CharacterDir: dir, YAMLPath: f, Backup: false})
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error = %q, want to contain \"validation failed\"", err.Error())
	}
}

func TestImport_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := Import(Options{CharacterDir: dir, YAMLPath: filepath.Join(dir, "missing.yml"), Backup: false})
	if err == nil {
		t.Fatal("expected error for missing YAML file")
	}
}

func TestImport_DryRun(t *testing.T) {
	dir := t.TempDir()
	src := testdata.CharDir()

	// Export real testdata to a YAML file
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	yamlPath := filepath.Join(dir, "macros.yml")
	if err := os.WriteFile(yamlPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	result, err := Import(Options{CharacterDir: destDir, YAMLPath: yamlPath, Backup: false, DryRun: true})
	if err != nil {
		t.Fatalf("Import dry-run: %v", err)
	}
	if result.BackupDir != "" {
		t.Errorf("DryRun should produce no backup dir, got %q", result.BackupDir)
	}
	if len(result.Sets) == 0 {
		t.Error("DryRun should report sets that would be written")
	}

	// No files should be written
	entries, _ := os.ReadDir(destDir)
	if len(entries) != 0 {
		t.Errorf("DryRun wrote %d files, want 0", len(entries))
	}
}

func TestImport_BackupCreated(t *testing.T) {
	charDir := t.TempDir()

	// Plant a fake .dat file to be backed up
	dummyPath := filepath.Join(charDir, "mcr.dat")
	if err := os.WriteFile(dummyPath, make([]byte, 7624), 0o644); err != nil {
		t.Fatal(err)
	}

	yamlPath := writeYAML(t, charDir, "m.yml", "version: 1\nbooks: {}\n")
	result, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: true})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if result.BackupDir == "" {
		t.Fatal("expected a backup dir path")
	}
	backed := filepath.Join(result.BackupDir, "mcr.dat")
	if _, err := os.Stat(backed); err != nil {
		t.Errorf("backed-up file not found at %s: %v", backed, err)
	}
}

func TestImport_RoundTrip(t *testing.T) {
	src := testdata.CharDir()

	// Export original macros to YAML
	srcDoc, err := export.FromCharacterDir(export.Options{CharacterDir: src, Character: "char"})
	if err != nil {
		t.Fatalf("export src: %v", err)
	}
	data, err := export.MarshalYAML(srcDoc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "macros.yml")
	if err := os.WriteFile(yamlPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	if _, err := Import(Options{CharacterDir: destDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Re-export from the destination and compare Books
	destDoc, err := export.FromCharacterDir(export.Options{CharacterDir: destDir, Character: "char"})
	if err != nil {
		t.Fatalf("export dest: %v", err)
	}

	if len(destDoc.Books) != len(srcDoc.Books) {
		t.Errorf("books count: got %d, want %d", len(destDoc.Books), len(srcDoc.Books))
	}
	for bookIdx, srcBook := range srcDoc.Books {
		destBook, ok := destDoc.Books[bookIdx]
		if !ok {
			t.Errorf("book %d missing from re-exported document", bookIdx)
			continue
		}
		if destBook.Name != srcBook.Name {
			t.Errorf("book %d name: got %q, want %q", bookIdx, destBook.Name, srcBook.Name)
		}
		for setIdx, srcSet := range srcBook.Sets {
			destSet, ok := destBook.Sets[setIdx]
			if !ok {
				t.Errorf("book %d set %d missing from re-exported document", bookIdx, setIdx)
				continue
			}
			compareRow(t, bookIdx, setIdx, "ctrl", srcSet.Ctrl, destSet.Ctrl)
			compareRow(t, bookIdx, setIdx, "alt", srcSet.Alt, destSet.Alt)
		}
	}
}

func compareRow(t *testing.T, book, set int, row string, src, dest map[int]export.Macro) {
	t.Helper()
	for key, sm := range src {
		dm, ok := dest[key]
		if !ok {
			t.Errorf("book %d set %d %s key %d: missing in dest", book, set, row, key)
			continue
		}
		if dm.Name != sm.Name {
			t.Errorf("book %d set %d %s key %d name: got %q, want %q", book, set, row, key, dm.Name, sm.Name)
		}
		if len(dm.Contents) != len(sm.Contents) {
			t.Errorf("book %d set %d %s key %d contents len: got %d, want %d", book, set, row, key, len(dm.Contents), len(sm.Contents))
			continue
		}
		for i, line := range sm.Contents {
			if dm.Contents[i] != line {
				t.Errorf("book %d set %d %s key %d contents[%d]: got %q, want %q", book, set, row, key, i, dm.Contents[i], line)
			}
		}
	}
}

func writeYAML(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
