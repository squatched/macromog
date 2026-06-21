package export_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/validate"
)

const datRoot = "../../data/dats"

func TestFromCharacterDir_Book33(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: filepath.Join(datRoot, "Book33"),
		Character:    "Book33Test",
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
		CharacterDir: filepath.Join(datRoot, "c75afe"),
	})
	if err != nil {
		t.Fatal(err)
	}
	book := doc.Books[6]
	set := book.Sets[10]
	if set.Ctrl[1].Name != "Ctrl1" {
		t.Errorf("ctrl[1].Name = %q", set.Ctrl[1].Name)
	}
	// SkpLines sits in alt slot 4 (YAML key 4); alt slot 1 is intentionally empty.
	if len(set.Alt[4].Contents) != 4 || set.Alt[4].Contents[3] != "Line 4" {
		t.Errorf("alt[4].Contents = %#v", set.Alt[4].Contents)
	}
}

func TestMarshalYAML_Validates(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: filepath.Join(datRoot, "Book33")})
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

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yml")
	if err := export.WriteFile(filepath.Join(datRoot, "Book33"), out, "test"); err != nil {
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