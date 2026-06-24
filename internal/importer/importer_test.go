package importer

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
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

// writeDummyDat creates a minimal valid DAT file at the given path.
func writeDummyDat(t *testing.T, dir, filename string) string {
	t.Helper()
	b := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(b[0:4], dat.MagicVersion)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImport_FullScopeClearing(t *testing.T) {
	charDir := t.TempDir()

	// Plant existing DAT files for book 1 set 1 and book 2 set 1.
	writeDummyDat(t, charDir, dat.MacroFileName(1, 1)) // B1S1
	writeDummyDat(t, charDir, dat.MacroFileName(2, 1)) // B2S1

	// Import a full-scope YAML that only mentions book 1, set 1.
	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
scope:
  level: full
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "Cure"
`)

	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// B1S1 should still exist (it was in the YAML).
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(1, 1))); err != nil {
		t.Errorf("B1S1 unexpectedly deleted: %v", err)
	}
	// B2S1 should be deleted (book 2 absent from full-scope YAML).
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(2, 1))); !os.IsNotExist(err) {
		t.Errorf("B2S1 should have been deleted (full scope clearing)")
	}
}

func TestImport_BookScopeClearing(t *testing.T) {
	charDir := t.TempDir()

	// Plant existing DAT files: book 1 sets 1 and 2; book 3 set 1 (out of scope).
	writeDummyDat(t, charDir, dat.MacroFileName(1, 1)) // B1S1 — in scope
	writeDummyDat(t, charDir, dat.MacroFileName(1, 2)) // B1S2 — in scope, absent from YAML
	writeDummyDat(t, charDir, dat.MacroFileName(3, 1)) // B3S1 — out of scope

	// Import a book-scope YAML that only mentions book 1, set 1.
	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
scope:
  level: book
  selections:
    - {book: 1}
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "Cure"
`)

	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// B1S1 should still exist (it was in the YAML).
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(1, 1))); err != nil {
		t.Errorf("B1S1 unexpectedly deleted: %v", err)
	}
	// B1S2 should be deleted (in scope, absent from YAML).
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(1, 2))); !os.IsNotExist(err) {
		t.Errorf("B1S2 should have been deleted (book scope clearing)")
	}
	// B3S1 should remain (out of scope).
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(3, 1))); err != nil {
		t.Errorf("B3S1 should be untouched (out of scope): %v", err)
	}
}

func TestImport_SetScopeClearing(t *testing.T) {
	charDir := t.TempDir()

	// Plant DAT files: B1S2 (in scope, no YAML content), B1S5 (out of scope).
	writeDummyDat(t, charDir, dat.MacroFileName(1, 2))
	writeDummyDat(t, charDir, dat.MacroFileName(1, 5))

	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
scope:
  level: set
  selections:
    - {book: 1, set: 2}
books: {}
`)

	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// B1S2 is in scope but absent from YAML → should be deleted.
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(1, 2))); !os.IsNotExist(err) {
		t.Errorf("B1S2 should have been deleted (set scope clearing)")
	}
	// B1S5 is out of scope → untouched.
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(1, 5))); err != nil {
		t.Errorf("B1S5 should be untouched (out of scope): %v", err)
	}
}

func TestImport_MacroScopeMerge(t *testing.T) {
	src := testdata.CharDir()

	// Export from testdata and import into a fresh dir first (full round-trip).
	srcDoc, err := export.FromCharacterDir(export.Options{CharacterDir: src})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	data, err := export.MarshalYAML(srcDoc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	tmpYAML := filepath.Join(t.TempDir(), "full.yml")
	if err := os.WriteFile(tmpYAML, data, 0o644); err != nil {
		t.Fatal(err)
	}
	destDir := t.TempDir()
	if _, err := Import(Options{CharacterDir: destDir, YAMLPath: tmpYAML, Backup: false}); err != nil {
		t.Fatalf("full import: %v", err)
	}

	// Now do a macro-scope import that only changes B6S10 ctrl 1.
	macroYAML := writeYAML(t, t.TempDir(), "macro.yml", `version: 1
scope:
  level: macro
  selections:
    - {book: 6, set: 10, type: ctrl, key: 1}
books:
  6:
    sets:
      10:
        ctrl:
          1:
            name: "Changed"
            contents:
              - /echo changed
`)

	if _, err := Import(Options{CharacterDir: destDir, YAMLPath: macroYAML, Scope: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{{Book: 6, Set: 10, Type: scope.TypeCtrl, Key: 1}}}, Backup: false}); err != nil {
		t.Fatalf("macro import: %v", err)
	}

	// Re-export and verify: ctrl 1 changed, ctrl 2+ and alt keys preserved.
	afterDoc, err := export.FromCharacterDir(export.Options{CharacterDir: destDir})
	if err != nil {
		t.Fatalf("re-export: %v", err)
	}
	b6 := afterDoc.Books[6]
	s10 := b6.Sets[10]
	if s10.Ctrl[1].Name != "Changed" {
		t.Errorf("ctrl 1 name = %q, want Changed", s10.Ctrl[1].Name)
	}
	// ctrl 2+ from the original should still be there.
	if srcSet := srcDoc.Books[6].Sets[10]; srcSet.Ctrl != nil {
		for key, m := range srcSet.Ctrl {
			if key == 1 {
				continue
			}
			if s10.Ctrl[key].Name != m.Name {
				t.Errorf("ctrl %d name: got %q, want %q (should be preserved)", key, s10.Ctrl[key].Name, m.Name)
			}
		}
	}
}

func TestImport_ScopeFiltersWrites(t *testing.T) {
	// When --scope is narrower than the YAML, only in-scope books are written.
	charDir := t.TempDir()

	// YAML has books 1 and 2, but scope restricts to book 1 only.
	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
scope:
  level: full
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "B1S1"
  2:
    sets:
      1:
        ctrl:
          1:
            name: "B2S1"
`)

	importScope := scope.Scope{
		Level:      scope.LevelBook,
		Selections: []scope.Selection{{Book: 1}},
	}
	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Scope: importScope, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Book 1 set 1 should be written.
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(1, 1))); err != nil {
		t.Errorf("B1S1 should have been written: %v", err)
	}
	// Book 2 set 1 should NOT be written (out of scope).
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(2, 1))); !os.IsNotExist(err) {
		t.Error("B2S1 should not be written when --scope restricts to book 1")
	}
}

func TestImport_FullScopeDeletesTitleFiles(t *testing.T) {
	charDir := t.TempDir()

	// Plant both .ttl files with some book titles.
	var titles [dat.MaxBooks]string
	titles[0] = "WHM75"   // book 1 → mcr.ttl
	titles[20] = "BLM90"  // book 21 → mcr_2.ttl
	if err := dat.WriteBookTitles(charDir, titles); err != nil {
		t.Fatalf("plant titles: %v", err)
	}

	// Full-scope import with no books → clears all titles.
	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
scope:
  level: full
books: {}
`)

	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Both .ttl files should be deleted since all titles are empty.
	for _, f := range []string{"mcr.ttl", "mcr_2.ttl"} {
		if _, err := os.Stat(filepath.Join(charDir, f)); !os.IsNotExist(err) {
			t.Errorf("%s should be deleted when all titles are cleared", f)
		}
	}
}

func TestImport_BookScopeDeletesOnlyRelevantTitleFile(t *testing.T) {
	charDir := t.TempDir()

	// Plant both .ttl files.
	var titles [dat.MaxBooks]string
	titles[0] = "WHM75"   // book 1 (in scope, will be cleared)
	titles[20] = "BLM90"  // book 21 (out of scope, should remain)
	if err := dat.WriteBookTitles(charDir, titles); err != nil {
		t.Fatalf("plant titles: %v", err)
	}

	// Book-scope import scoped to book 1 with no book 1 content.
	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
scope:
  level: book
  selections:
    - {book: 1}
books: {}
`)

	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// mcr.ttl (books 1-20): book 1 cleared, all empty → file deleted.
	if _, err := os.Stat(filepath.Join(charDir, "mcr.ttl")); !os.IsNotExist(err) {
		t.Error("mcr.ttl should be deleted when all books 1-20 have empty titles")
	}
	// mcr_2.ttl (books 21-40): book 21 untouched → file remains.
	if _, err := os.Stat(filepath.Join(charDir, "mcr_2.ttl")); err != nil {
		t.Errorf("mcr_2.ttl should remain (book 21 title preserved): %v", err)
	}
}

func TestImport_LegacyNoScope_WriteOnly(t *testing.T) {
	charDir := t.TempDir()

	// Plant a DAT file that should NOT be cleared (legacy YAML = write-only).
	writeDummyDat(t, charDir, dat.MacroFileName(2, 1))

	// YAML with no scope field (legacy).
	yamlPath := writeYAML(t, charDir, "macros.yml", `version: 1
books:
  1:
    sets:
      1:
        ctrl:
          1:
            name: "Cure"
`)

	if _, err := Import(Options{CharacterDir: charDir, YAMLPath: yamlPath, Backup: false}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// B2S1 should still exist — legacy mode does no clearing.
	if _, err := os.Stat(filepath.Join(charDir, dat.MacroFileName(2, 1))); err != nil {
		t.Errorf("B2S1 should not be cleared in legacy write-only mode: %v", err)
	}
}
