package template_test

import (
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
	tmpl "github.com/squatched/macromog/internal/template"
	"github.com/squatched/macromog/internal/validate"
)

func TestGenerate_FullScope(t *testing.T) {
	doc := tmpl.Generate(scope.Full(), "")
	if doc.Scope.Level != scope.LevelFull {
		t.Errorf("scope level = %q, want full", doc.Scope.Level)
	}
	if doc.ExportedAt != "" {
		t.Error("template should not have exported_at")
	}
	if len(doc.Books) != dat.MaxBooks {
		t.Errorf("books = %d, want %d", len(doc.Books), dat.MaxBooks)
	}
	// Every book should have all 10 sets with 10 ctrl + 10 alt each.
	for bookIdx, b := range doc.Books {
		if len(b.Sets) != dat.SetsPerBook {
			t.Errorf("book %d: sets = %d, want %d", bookIdx, len(b.Sets), dat.SetsPerBook)
			continue
		}
		for setIdx, s := range b.Sets {
			if len(s.Ctrl) != dat.SetsPerBook {
				t.Errorf("book %d set %d: ctrl count = %d, want %d", bookIdx, setIdx, len(s.Ctrl), dat.SetsPerBook)
			}
			if len(s.Alt) != dat.SetsPerBook {
				t.Errorf("book %d set %d: alt count = %d, want %d", bookIdx, setIdx, len(s.Alt), dat.SetsPerBook)
			}
			if m, ok := s.Ctrl[1]; ok && len(m.Contents) != dat.LineCount {
				t.Errorf("book %d set %d ctrl 1: contents len = %d, want %d", bookIdx, setIdx, len(m.Contents), dat.LineCount)
			}
		}
	}
}

func TestGenerate_Character(t *testing.T) {
	doc := tmpl.Generate(scope.Full(), "Squatched")
	if doc.Character != "Squatched" {
		t.Errorf("character = %q, want Squatched", doc.Character)
	}
	doc2 := tmpl.Generate(scope.Full(), "")
	if doc2.Character != "" {
		t.Errorf("character should be empty when not provided, got %q", doc2.Character)
	}
}

func TestGenerate_BookScope(t *testing.T) {
	sc := scope.Scope{
		Level:      scope.LevelBook,
		Selections: []scope.Selection{{Book: 1}, {Book: 5}},
	}
	doc := tmpl.Generate(sc, "")
	if len(doc.Books) != 2 {
		t.Errorf("books = %d, want 2", len(doc.Books))
	}
	if _, ok := doc.Books[1]; !ok {
		t.Error("book 1 missing")
	}
	if _, ok := doc.Books[5]; !ok {
		t.Error("book 5 missing")
	}
	if _, ok := doc.Books[2]; ok {
		t.Error("book 2 should not be present")
	}
}

func TestGenerate_SetScope(t *testing.T) {
	sc := scope.Scope{
		Level:      scope.LevelSet,
		Selections: []scope.Selection{{Book: 1, Set: 3}, {Book: 1, Set: 7}},
	}
	doc := tmpl.Generate(sc, "")
	b1, ok := doc.Books[1]
	if !ok {
		t.Fatal("book 1 missing")
	}
	if len(b1.Sets) != 2 {
		t.Errorf("book 1 sets = %d, want 2", len(b1.Sets))
	}
	if _, ok := b1.Sets[3]; !ok {
		t.Error("set 3 missing")
	}
	if _, ok := b1.Sets[7]; !ok {
		t.Error("set 7 missing")
	}
	if _, ok := b1.Sets[1]; ok {
		t.Error("set 1 should not be present")
	}
}

func TestGenerate_MacroScope(t *testing.T) {
	sc := scope.Scope{
		Level: scope.LevelMacro,
		Selections: []scope.Selection{
			{Book: 1, Set: 3, Type: scope.TypeAlt, Key: 1},
			{Book: 1, Set: 3, Type: scope.TypeCtrl, Key: 2},
		},
	}
	doc := tmpl.Generate(sc, "")
	b1, ok := doc.Books[1]
	if !ok {
		t.Fatal("book 1 missing")
	}
	s3 := b1.Sets[3]
	if len(s3.Ctrl) != 1 {
		t.Errorf("ctrl keys = %d, want 1", len(s3.Ctrl))
	}
	if len(s3.Alt) != 1 {
		t.Errorf("alt keys = %d, want 1", len(s3.Alt))
	}
	if _, ok := s3.Ctrl[2]; !ok {
		t.Error("ctrl 2 missing")
	}
	if _, ok := s3.Alt[1]; !ok {
		t.Error("alt 1 missing")
	}
}

func TestGenerate_EmptyContents(t *testing.T) {
	sc := scope.Scope{
		Level:      scope.LevelSet,
		Selections: []scope.Selection{{Book: 1, Set: 1}},
	}
	doc := tmpl.Generate(sc, "")
	m := doc.Books[1].Sets[1].Ctrl[1]
	if len(m.Contents) != dat.LineCount {
		t.Errorf("contents len = %d, want %d", len(m.Contents), dat.LineCount)
	}
	for i, line := range m.Contents {
		if line != "" {
			t.Errorf("contents[%d] = %q, want empty string", i, line)
		}
	}
}

func TestGenerate_ValidatesClean(t *testing.T) {
	sc := scope.Scope{
		Level:      scope.LevelBook,
		Selections: []scope.Selection{{Book: 1}},
	}
	doc := tmpl.Generate(sc, "char")
	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	if errs := validate.Validate(data); len(errs) > 0 {
		t.Errorf("template YAML failed validation: %v", errs)
	}
	// Template should not have exported_at in output.
	if strings.Contains(string(data), "exported_at") {
		t.Error("template output should not contain exported_at")
	}
}
