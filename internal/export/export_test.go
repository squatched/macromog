package export_test

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
	"github.com/squatched/macromog/internal/validate"
)

func TestFromCharacterDir_Book33(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Character:    "testchar",
	})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 1 {
		t.Errorf("version = %d", doc.Version)
	}
	book, ok := doc.Books[33]
	if !ok {
		t.Fatal("missing book 33")
	}
	set, ok := book.Sets[1]
	if !ok {
		t.Fatal("missing set 1")
	}
	m, ok := set.Ctrl[1]
	if !ok || m.Name != "B33S1" {
		t.Fatalf("ctrl[1] = %#v", set.Ctrl[1])
	}
}

func TestFromCharacterDir_StructTest(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	book := doc.Books[6]
	set := book.Sets[10]
	if set.Ctrl[1].Name != "Ctrl1" {
		t.Errorf("ctrl[1].Name = %q", set.Ctrl[1].Name)
	}
	if len(set.Alt[4].Contents) != 4 || set.Alt[4].Contents[3] != "Line 4" {
		t.Errorf("alt[4].Contents = %#v", set.Alt[4].Contents)
	}
}

func TestMarshalYAML_Validates(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: testdata.CharDir()})
	if err != nil {
		t.Fatal(err)
	}
	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	if errs := validate.Validate(data); len(errs) > 0 {
		t.Fatalf("validation errors: %v", errs)
	}
}

func TestFromCharacterDir_MissingDir(t *testing.T) {
	_, err := export.FromCharacterDir(export.Options{CharacterDir: "/nonexistent/char"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFromCharacterDir_EmptyCharacter(t *testing.T) {
	dir := t.TempDir()
	blank := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(blank[0:4], dat.MagicVersion)
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), blank, 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: dir,
		Character:    "newbie",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Character != "newbie" {
		t.Errorf("character = %q, want newbie", doc.Character)
	}
	if len(doc.Books) != 0 {
		t.Errorf("books = %d, want empty document", len(doc.Books))
	}

	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "version: 1") {
		t.Errorf("expected version in output: %s", data)
	}
	if strings.Contains(string(data), "books:") {
		t.Errorf("sparse empty export should omit books key: %s", data)
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yml")
	if err := export.WriteFile(export.Options{CharacterDir: testdata.CharDir(), Character: "test"}, out); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "B33S1") {
		t.Errorf("output missing macro name: %s", data)
	}
}

func TestFromCharacterDir_ScopeFieldAlwaysPresent(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: testdata.CharDir()})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Scope.Level != scope.LevelFull {
		t.Errorf("default scope level = %q, want full", doc.Scope.Level)
	}
	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "level: full") {
		t.Errorf("marshaled YAML missing scope field:\n%s", data)
	}
}

func TestFromCharacterDir_BookScope(t *testing.T) {
	sc := scope.Scope{
		Level:      scope.LevelBook,
		Selections: []scope.Selection{{Book: 33}},
	}
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Scope:        sc,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Only book 33 should appear.
	if _, ok := doc.Books[33]; !ok {
		t.Error("book 33 missing from book-scoped export")
	}
	for bk := range doc.Books {
		if bk != 33 {
			t.Errorf("book-scoped export contains unexpected book %d", bk)
		}
	}
	if doc.Scope.Level != scope.LevelBook {
		t.Errorf("scope.Level = %q, want book", doc.Scope.Level)
	}
}

func TestFromCharacterDir_SetScope(t *testing.T) {
	sc := scope.Scope{
		Level:      scope.LevelSet,
		Selections: []scope.Selection{{Book: 6, Set: 10}},
	}
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Scope:        sc,
	})
	if err != nil {
		t.Fatal(err)
	}
	b6, ok := doc.Books[6]
	if !ok {
		t.Fatal("book 6 missing from set-scoped export")
	}
	if _, ok := b6.Sets[10]; !ok {
		t.Error("set 10 missing from set-scoped export")
	}
	for setIdx := range b6.Sets {
		if setIdx != 10 {
			t.Errorf("set-scoped export contains unexpected set %d in book 6", setIdx)
		}
	}
}