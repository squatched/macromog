package export_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
	"github.com/squatched/macromog/internal/validate"
	"gopkg.in/yaml.v3"
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
	s := string(data)
	if !strings.Contains(s, "B33S1") {
		t.Errorf("output missing macro name: %s", data)
	}
	// WriteFile must never emit bare double-quoted empty strings; they mislead
	// users into wrapping macro content in outer quotes.
	if strings.Contains(s, `- ""`) {
		t.Errorf("WriteFile output contains raw empty-string items:\n%s", data)
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

func TestFromCharacterDir_Dense_IncludesAllBooks(t *testing.T) {
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Dense:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Books) != dat.MaxBooks {
		t.Errorf("dense export: books = %d, want %d", len(doc.Books), dat.MaxBooks)
	}
	for bookIdx, b := range doc.Books {
		if len(b.Sets) != dat.SetsPerBook {
			t.Errorf("book %d: sets = %d, want %d", bookIdx, len(b.Sets), dat.SetsPerBook)
		}
		for setIdx, s := range b.Sets {
			if len(s.Ctrl) != dat.SetsPerBook {
				t.Errorf("book %d set %d: ctrl count = %d, want %d", bookIdx, setIdx, len(s.Ctrl), dat.SetsPerBook)
			}
			if len(s.Alt) != dat.SetsPerBook {
				t.Errorf("book %d set %d: alt count = %d, want %d", bookIdx, setIdx, len(s.Alt), dat.SetsPerBook)
			}
		}
	}
}

func TestFromCharacterDir_Dense_IncludesEmptyMacros(t *testing.T) {
	// Testdata has content only in specific books; dense export must include
	// macros from books that have no .dat file on disk (treated as empty).
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Dense:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Book 1 has no .dat file in testdata; all its slots should be present
	// with non-nil Contents (empty slice) in dense mode.
	b1, ok := doc.Books[1]
	if !ok {
		t.Fatal("book 1 missing from dense export")
	}
	s1, ok := b1.Sets[1]
	if !ok {
		t.Fatal("book 1 set 1 missing from dense export")
	}
	m, ok := s1.Ctrl[1]
	if !ok {
		t.Fatal("ctrl[1] missing from dense export of book 1 set 1")
	}
	if m.Contents == nil {
		t.Error("dense export: empty macro should have non-nil Contents, got nil")
	}
}

func TestFromCharacterDir_Dense_PreservesContent(t *testing.T) {
	// Dense export must not lose content that exists in .dat files.
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Character:    "testchar",
		Dense:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	book, ok := doc.Books[33]
	if !ok {
		t.Fatal("book 33 missing from dense export")
	}
	set, ok := book.Sets[1]
	if !ok {
		t.Fatal("set 1 missing from dense export of book 33")
	}
	m, ok := set.Ctrl[1]
	if !ok || m.Name != "B33S1" {
		t.Fatalf("ctrl[1] = %#v, want Name=B33S1", m)
	}
}

func TestFromCharacterDir_Dense_Scope(t *testing.T) {
	// Dense export respects scope: only the requested book/set should appear.
	sc := scope.Scope{
		Level:      scope.LevelSet,
		Selections: []scope.Selection{{Book: 1, Set: 3}},
	}
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: testdata.CharDir(),
		Dense:        true,
		Scope:        sc,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Books) != 1 {
		t.Errorf("dense scoped export: books = %d, want 1", len(doc.Books))
	}
	b1, ok := doc.Books[1]
	if !ok {
		t.Fatal("book 1 missing from dense scoped export")
	}
	if len(b1.Sets) != 1 {
		t.Errorf("book 1: sets = %d, want 1", len(b1.Sets))
	}
	if _, ok := b1.Sets[3]; !ok {
		t.Error("set 3 missing from dense scoped export")
	}
}

func TestMarshalYAML_NameOnlyMacro(t *testing.T) {
	// A macro with a name but no content lines.
	// Sparse: only the name key; no contents key at all.
	// Dense:  name key plus contents with 6 numbered comment placeholders.
	dir := t.TempDir()
	set := dat.MacroSet{}
	set.Ctrl[0].Name = "Cure"
	// all Contents remain ""
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), dat.EncodeMacroSet(set), 0o644); err != nil {
		t.Fatal(err)
	}

	sc := scope.Scope{
		Level:      scope.LevelSet,
		Selections: []scope.Selection{{Book: 1, Set: 1}},
	}

	for _, tc := range []struct {
		label string
		dense bool
	}{{"sparse", false}, {"dense", true}} {
		t.Run(tc.label, func(t *testing.T) {
			doc, err := export.FromCharacterDir(export.Options{
				CharacterDir: dir,
				Dense:        tc.dense,
				Scope:        sc,
			})
			if err != nil {
				t.Fatal(err)
			}
			data, err := export.MarshalYAMLWithPlaceholders(doc)
			if err != nil {
				t.Fatal(err)
			}
			s := string(data)

			if !strings.Contains(s, "name: Cure") {
				t.Fatalf("macro name missing from %s export:\n%s", tc.label, s)
			}

			if tc.dense {
				for i := 1; i <= 6; i++ {
					ph := fmt.Sprintf("# Macro Line %d", i)
					if !strings.Contains(s, ph) {
						t.Errorf("dense name-only macro: missing %q:\n%s", ph, s)
					}
				}
				if strings.Contains(s, `contents: []`) {
					t.Errorf("dense name-only macro must not produce 'contents: []':\n%s", s)
				}
			} else {
				if strings.Contains(s, "contents:") {
					t.Errorf("sparse name-only macro must not have a contents key:\n%s", s)
				}
			}
		})
	}
}

func TestMarshalYAML_Dense_NoDoubleQuotedEmptyLines(t *testing.T) {
	// Dense export of an empty character dir must not produce "" for empty
	// macro content lines.
	dir := t.TempDir()
	blank := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(blank[0:4], dat.MagicVersion)
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), blank, 0o644); err != nil {
		t.Fatal(err)
	}

	sc := scope.Scope{
		Level:      scope.LevelSet,
		Selections: []scope.Selection{{Book: 1, Set: 1}},
	}
	doc, err := export.FromCharacterDir(export.Options{
		CharacterDir: dir,
		Dense:        true,
		Scope:        sc,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := export.MarshalYAMLWithPlaceholders(doc)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "# Macro Line 1") {
		t.Errorf("dense export YAML should use comment placeholders for empty slots:\n%s", s)
	}
	if errs := validate.Validate(data); len(errs) > 0 {
		t.Fatalf("dense export YAML failed validation: %v", errs)
	}
}
// makeMacroSet writes a .dat file into a temp dir with one Ctrl macro at slot 0
// using the given name and exactly the provided content lines (padded to 6).
func makeMacroSet(t *testing.T, name string, contentLines ...string) string {
	t.Helper()
	dir := t.TempDir()
	set := dat.MacroSet{}
	set.Ctrl[0].Name = name
	for i, l := range contentLines {
		if i < dat.LineCount {
			set.Ctrl[0].Contents[i] = l
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), dat.EncodeMacroSet(set), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// marshalScope exports dir with a B1S1 scope + given dense flag and returns
// the post-processed YAML string.
func marshalScope(t *testing.T, dir string, dense bool) string {
	t.Helper()
	sc := scope.Scope{Level: scope.LevelSet, Selections: []scope.Selection{{Book: 1, Set: 1}}}
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: dir, Dense: dense, Scope: sc})
	if err != nil {
		t.Fatal(err)
	}
	data, err := export.MarshalYAMLWithPlaceholders(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMarshalYAML_InteriorBlank(t *testing.T) {
	// Macro with content on line 1 and line 3; line 2 is blank.
	dir := makeMacroSet(t, "Provoke", "/echo attack", "", `/ja "Provoke" <t>`)

	t.Run("sparse", func(t *testing.T) {
		s := marshalScope(t, dir, false)
		if !strings.Contains(s, "- # Macro Line 2") {
			t.Errorf("sparse: interior blank not replaced with comment:\n%s", s)
		}
		if strings.Contains(s, "# Macro Line 4") {
			t.Errorf("sparse: trailing placeholder lines must not appear:\n%s", s)
		}
		if strings.Contains(s, `- ""`) {
			t.Errorf("sparse: raw empty-string items must not appear:\n%s", s)
		}
	})

	t.Run("dense", func(t *testing.T) {
		s := marshalScope(t, dir, true)
		for _, ph := range []string{"# Macro Line 2", "# Macro Line 4", "# Macro Line 5", "# Macro Line 6"} {
			if !strings.Contains(s, ph) {
				t.Errorf("dense: missing %q:\n%s", ph, s)
			}
		}
		if !strings.Contains(s, "/echo attack") || !strings.Contains(s, `/ja "Provoke" <t>`) {
			t.Errorf("dense: real content missing:\n%s", s)
		}
		if strings.Contains(s, `- ""`) {
			t.Errorf("dense: raw empty-string items must not appear:\n%s", s)
		}
	})
}

func TestMarshalYAML_AllContentLines_Dense(t *testing.T) {
	// When all 6 content lines are non-empty the post-processor must not
	// replace or inject any comment in the populated macro's contents block.
	// Dense mode also emits the other empty slots — check only the "Full" block.
	dir := makeMacroSet(t, "Full",
		"/echo 1", "/echo 2", "/echo 3", "/echo 4", "/echo 5", "/echo 6")
	s := marshalScope(t, dir, true)

	// Extract the section belonging to "Full" (between "name: Full" and the
	// next sibling key or end of string) and assert no placeholder appears there.
	block := sectionAfter(s, "name: Full")
	if strings.Contains(block, "# Macro Line") {
		t.Errorf("all-content macro: placeholder comments must not appear in its contents block:\n%s", block)
	}
	for i := 1; i <= 6; i++ {
		if !strings.Contains(s, fmt.Sprintf("/echo %d", i)) {
			t.Errorf("all-content macro: line %d missing from dense output:\n%s", i, s)
		}
	}
}

// sectionAfter returns the lines starting from the first occurrence of marker
// through the next line that starts at the same or lower indentation level as
// marker (exclusive), or end of string.
func sectionAfter(s, marker string) string {
	lines := strings.Split(s, "\n")
	start := -1
	markerIndent := 0
	for i, l := range lines {
		if strings.Contains(l, marker) {
			start = i
			markerIndent = len(l) - len(strings.TrimLeft(l, " "))
			break
		}
	}
	if start < 0 {
		return ""
	}
	var out []string
	for _, l := range lines[start:] {
		indent := len(l) - len(strings.TrimLeft(l, " "))
		if len(out) > 0 && l != "" && indent <= markerIndent {
			break
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func TestMarshalYAML_ContentLastLineOnly(t *testing.T) {
	// Content only on line 6; lines 1-5 are blank.
	dir := makeMacroSet(t, "Stab", "", "", "", "", "", `/ws "Savage Blade" <t>`)

	t.Run("sparse", func(t *testing.T) {
		s := marshalScope(t, dir, false)
		for i := 1; i <= 5; i++ {
			if !strings.Contains(s, fmt.Sprintf("# Macro Line %d", i)) {
				t.Errorf("sparse last-line: missing # Macro Line %d:\n%s", i, s)
			}
		}
		if !strings.Contains(s, `/ws "Savage Blade" <t>`) {
			t.Errorf("sparse last-line: content on line 6 missing:\n%s", s)
		}
		if strings.Contains(s, `- ""`) {
			t.Errorf("sparse last-line: raw empty-string items must not appear:\n%s", s)
		}
	})

	t.Run("dense", func(t *testing.T) {
		s := marshalScope(t, dir, true)
		for i := 1; i <= 5; i++ {
			if !strings.Contains(s, fmt.Sprintf("# Macro Line %d", i)) {
				t.Errorf("dense last-line: missing # Macro Line %d:\n%s", i, s)
			}
		}
		if strings.Contains(s, "# Macro Line 7") {
			t.Errorf("dense last-line: phantom line 7 must not appear:\n%s", s)
		}
		if !strings.Contains(s, `/ws "Savage Blade" <t>`) {
			t.Errorf("dense last-line: content on line 6 missing:\n%s", s)
		}
	})
}

// ---------------------------------------------------------------------------
// Scope serialization: nil Key ↔ absent "key:" field round-trip.
// ---------------------------------------------------------------------------

func TestMarshalYAML_ScopeSelectionKeyField(t *testing.T) {
	kp := func(n int) *int { return &n }
	tests := []struct {
		name       string
		sel        scope.Selection
		wantKey    bool
		wantKeyVal string
	}{
		{
			name:    "ctrl wildcard (nil Key): no key field",
			sel:     scope.Selection{Book: 1, Set: 1, Type: scope.TypeCtrl},
			wantKey: false,
		},
		{
			name:    "alt wildcard (nil Key): no key field",
			sel:     scope.Selection{Book: 1, Set: 1, Type: scope.TypeAlt},
			wantKey: false,
		},
		{
			name:       "ctrl key 0: key: 0 present",
			sel:        scope.Selection{Book: 1, Set: 1, Type: scope.TypeCtrl, Key: kp(0)},
			wantKey:    true,
			wantKeyVal: "0",
		},
		{
			name:       "ctrl key 9: key: 9 present",
			sel:        scope.Selection{Book: 1, Set: 1, Type: scope.TypeCtrl, Key: kp(9)},
			wantKey:    true,
			wantKeyVal: "9",
		},
		{
			name:       "alt key 0: key: 0 present",
			sel:        scope.Selection{Book: 1, Set: 1, Type: scope.TypeAlt, Key: kp(0)},
			wantKey:    true,
			wantKeyVal: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := export.Document{
				Version: 1,
				Scope:   scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{tt.sel}},
			}
			data, err := export.MarshalYAML(doc)
			if err != nil {
				t.Fatalf("MarshalYAML: %v", err)
			}
			yamlStr := string(data)
			if tt.wantKey {
				if !strings.Contains(yamlStr, "key: "+tt.wantKeyVal) {
					t.Errorf("expected 'key: %s' in YAML:\n%s", tt.wantKeyVal, yamlStr)
				}
			} else {
				if strings.Contains(yamlStr, "key:") {
					t.Errorf("wildcard selection must have no 'key:' field:\n%s", yamlStr)
				}
			}
		})
	}
}

// TestMarshalYAML_ScopeKeyRoundTrip verifies the critical invariant:
// nil Key → absent "key:" field → nil Key after unmarshal.
// This is distinct from Key: &0, which must survive as key: 0.
func TestMarshalYAML_ScopeKeyRoundTrip(t *testing.T) {
	kp := func(n int) *int { return &n }
	sc := scope.Scope{
		Level: scope.LevelMacro,
		Selections: []scope.Selection{
			{Book: 1, Set: 1, Type: scope.TypeCtrl},        // wildcard
			{Book: 1, Set: 1, Type: scope.TypeAlt, Key: kp(0)}, // key 0
			{Book: 1, Set: 1, Type: scope.TypeAlt, Key: kp(9)}, // key 9
		},
	}
	doc := export.Document{Version: 1, Scope: sc}
	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}

	var parsed export.Document
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	sels := parsed.Scope.Selections
	if len(sels) != 3 {
		t.Fatalf("got %d selections, want 3", len(sels))
	}
	if sels[0].Key != nil {
		t.Errorf("sel[0] (wildcard): Key = %v, want nil", sels[0].Key)
	}
	if sels[1].Key == nil {
		t.Error("sel[1] (key 0): Key is nil, want &0")
	} else if *sels[1].Key != 0 {
		t.Errorf("sel[1] (key 0): *Key = %d, want 0", *sels[1].Key)
	}
	if sels[2].Key == nil {
		t.Error("sel[2] (key 9): Key is nil, want &9")
	} else if *sels[2].Key != 9 {
		t.Errorf("sel[2] (key 9): *Key = %d, want 9", *sels[2].Key)
	}
}

// TestMarshalYAML_TwoTypeWildcardScope verifies the C,A pattern serializes
// as two selections with no key field in either.
func TestMarshalYAML_TwoTypeWildcardScope(t *testing.T) {
	sc := scope.Scope{
		Level: scope.LevelMacro,
		Selections: []scope.Selection{
			{Book: 40, Set: 10, Type: scope.TypeCtrl},
			{Book: 40, Set: 10, Type: scope.TypeAlt},
		},
	}
	doc := export.Document{Version: 1, Scope: sc}
	data, err := export.MarshalYAML(doc)
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}
	yamlStr := string(data)
	if strings.Contains(yamlStr, "key:") {
		t.Errorf("two-type wildcard scope must produce no key field:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "type: ctrl") {
		t.Errorf("expected 'type: ctrl' in output:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "type: alt") {
		t.Errorf("expected 'type: alt' in output:\n%s", yamlStr)
	}
}
