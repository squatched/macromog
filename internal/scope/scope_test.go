package scope_test

import (
	"testing"

	"github.com/squatched/macromog/internal/scope"
)

func TestParseSelectorsEmpty(t *testing.T) {
	s, err := scope.ParseSelectors(nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != scope.LevelFull {
		t.Errorf("empty flags: got level %q, want full", s.Level)
	}
	if len(s.Selections) != 0 {
		t.Errorf("full scope should have no selections, got %v", s.Selections)
	}
}

func TestParseSelectorsWildcard(t *testing.T) {
	for _, flag := range []string{"*", "B*", "b*"} {
		s, err := scope.ParseSelectors([]string{flag})
		if err != nil {
			t.Fatalf("%q: %v", flag, err)
		}
		if s.Level != scope.LevelFull {
			t.Errorf("%q: got level %q, want full", flag, s.Level)
		}
	}
}

func TestParseSelectorsBook(t *testing.T) {
	tests := []struct {
		flag string
		want []scope.Selection
	}{
		{"B1", []scope.Selection{{Book: 1}}},
		{"B5", []scope.Selection{{Book: 5}}},
		{"B1,3,5", []scope.Selection{{Book: 1}, {Book: 3}, {Book: 5}}},
		{"B1-3", []scope.Selection{{Book: 1}, {Book: 2}, {Book: 3}}},
	}
	for _, tt := range tests {
		s, err := scope.ParseSelectors([]string{tt.flag})
		if err != nil {
			t.Fatalf("%q: %v", tt.flag, err)
		}
		if s.Level != scope.LevelBook {
			t.Errorf("%q: level = %q, want book", tt.flag, s.Level)
		}
		if !selectionsEqual(s.Selections, tt.want) {
			t.Errorf("%q: selections = %v, want %v", tt.flag, s.Selections, tt.want)
		}
	}
}

func TestParseSelectorsBookWildcardCollapsesToBook(t *testing.T) {
	// B1S* = all sets in book 1 = book-level authority for book 1.
	s, err := scope.ParseSelectors([]string{"B1S*"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != scope.LevelBook {
		t.Errorf("B1S*: level = %q, want book", s.Level)
	}
	if len(s.Selections) != 1 || s.Selections[0].Book != 1 {
		t.Errorf("B1S*: selections = %v, want [{Book:1}]", s.Selections)
	}
}

func TestParseSelectorsSet(t *testing.T) {
	tests := []struct {
		flag string
		want []scope.Selection
	}{
		{"B1S3", []scope.Selection{{Book: 1, Set: 3}}},
		{"B1S2,4", []scope.Selection{{Book: 1, Set: 2}, {Book: 1, Set: 4}}},
		{"B1S1-3", []scope.Selection{{Book: 1, Set: 1}, {Book: 1, Set: 2}, {Book: 1, Set: 3}}},
	}
	for _, tt := range tests {
		s, err := scope.ParseSelectors([]string{tt.flag})
		if err != nil {
			t.Fatalf("%q: %v", tt.flag, err)
		}
		if s.Level != scope.LevelSet {
			t.Errorf("%q: level = %q, want set", tt.flag, s.Level)
		}
		if !selectionsEqual(s.Selections, tt.want) {
			t.Errorf("%q: selections = %v, want %v", tt.flag, s.Selections, tt.want)
		}
	}
}

func TestParseSelectionsMacro(t *testing.T) {
	tests := []struct {
		flags []string
		want  []scope.Selection
	}{
		{
			[]string{"B1S3C2"},
			[]scope.Selection{{Book: 1, Set: 3, Type: scope.TypeCtrl, Key: keyPtr(2)}},
		},
		{
			[]string{"B1S3A1"},
			[]scope.Selection{{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(1)}},
		},
		{
			// Comma sibling: A1 then C2, same B1S3 context.
			[]string{"B1S3A1,C2"},
			[]scope.Selection{
				{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(1)},
				{Book: 1, Set: 3, Type: scope.TypeCtrl, Key: keyPtr(2)},
			},
		},
		{
			// Bare number sibling: A1,3 = alt keys 1 and 3.
			[]string{"B1S3A1,3"},
			[]scope.Selection{
				{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(1)},
				{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(3)},
			},
		},
		{
			// Multiple --scope flags.
			[]string{"B1S3A1", "B5S2C4"},
			[]scope.Selection{
				{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(1)},
				{Book: 5, Set: 2, Type: scope.TypeCtrl, Key: keyPtr(4)},
			},
		},
	}
	for _, tt := range tests {
		s, err := scope.ParseSelectors(tt.flags)
		if err != nil {
			t.Fatalf("%v: %v", tt.flags, err)
		}
		if s.Level != scope.LevelMacro {
			t.Errorf("%v: level = %q, want macro", tt.flags, s.Level)
		}
		if !selectionsEqual(s.Selections, tt.want) {
			t.Errorf("%v: selections = %v, want %v", tt.flags, s.Selections, tt.want)
		}
	}
}

func TestParseSelectorsCtrlWildcard(t *testing.T) {
	s, err := scope.ParseSelectors([]string{"B1S3C*"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != scope.LevelMacro {
		t.Errorf("level = %q, want macro", s.Level)
	}
	if len(s.Selections) != 1 {
		t.Fatalf("want 1 ctrl selection (wildcard), got %d: %v", len(s.Selections), s.Selections)
	}
	sel := s.Selections[0]
	if sel.Book != 1 || sel.Set != 3 || sel.Type != scope.TypeCtrl || sel.Key != nil {
		t.Errorf("selection = %v, want {Book:1, Set:3, Type:ctrl, Key:nil}", sel)
	}
}

func TestParseSelectorsAltWildcard(t *testing.T) {
	s, err := scope.ParseSelectors([]string{"B1S3A*"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != scope.LevelMacro {
		t.Errorf("level = %q, want macro", s.Level)
	}
	if len(s.Selections) != 1 {
		t.Fatalf("want 1 alt selection (wildcard), got %d: %v", len(s.Selections), s.Selections)
	}
	sel := s.Selections[0]
	if sel.Book != 1 || sel.Set != 3 || sel.Type != scope.TypeAlt || sel.Key != nil {
		t.Errorf("selection = %v, want {Book:1, Set:3, Type:alt, Key:nil}", sel)
	}
}

func TestParseSelectorsImplicitWildcard(t *testing.T) {
	// Bare C or A with no numspec is equivalent to C* / A*.
	tests := []struct {
		flag string
		want []scope.Selection
	}{
		// Single type, no numspec
		{"B1S3C", []scope.Selection{{Book: 1, Set: 3, Type: scope.TypeCtrl}}},
		{"B1S3A", []scope.Selection{{Book: 1, Set: 3, Type: scope.TypeAlt}}},
		// Both types — the primary use case (C,A = all ctrl and all alt)
		{"B1S3C,A", []scope.Selection{
			{Book: 1, Set: 3, Type: scope.TypeCtrl},
			{Book: 1, Set: 3, Type: scope.TypeAlt},
		}},
		// Reversed order
		{"B1S3A,C", []scope.Selection{
			{Book: 1, Set: 3, Type: scope.TypeAlt},
			{Book: 1, Set: 3, Type: scope.TypeCtrl},
		}},
		// Mix of explicit wildcard and implicit wildcard
		{"B1S3C*,A", []scope.Selection{
			{Book: 1, Set: 3, Type: scope.TypeCtrl},
			{Book: 1, Set: 3, Type: scope.TypeAlt},
		}},
	}
	for _, tt := range tests {
		s, err := scope.ParseSelectors([]string{tt.flag})
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", tt.flag, err)
		}
		if s.Level != scope.LevelMacro {
			t.Errorf("%q: level = %q, want macro", tt.flag, s.Level)
		}
		if !selectionsEqual(s.Selections, tt.want) {
			t.Errorf("%q: selections = %v, want %v", tt.flag, s.Selections, tt.want)
		}
	}
}

func TestParseSelectionsMixedLevelsError(t *testing.T) {
	_, err := scope.ParseSelectors([]string{"B1", "B2S3"})
	if err == nil {
		t.Error("expected error for mixed levels, got nil")
	}
}

func TestParseSelectorsErrors(t *testing.T) {
	bad := []string{
		"B41",       // book out of range
		"B1S11",     // set out of range
		"B1S3C10",   // ctrl key out of range
		"S3",        // S without B
		"C2",        // C without S
		"A2",        // A without S
		"B1,2S3",    // multiple books before S
		"B1S2,3C4",  // multiple sets before C
		"B1X",       // unknown component
		"",          // empty
		"B1S3C,3",   // bare C before digit sibling — inconsistent with B,1 rule
		"B1S3A,3",   // bare A before digit sibling — same rule
	}
	for _, flag := range bad {
		_, err := scope.ParseSelectors([]string{flag})
		if err == nil {
			t.Errorf("expected error for %q, got nil", flag)
		}
	}
}

func TestContainsBook(t *testing.T) {
	full := scope.Full()
	book := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}, {Book: 5}}}
	set := scope.Scope{Level: scope.LevelSet, Selections: []scope.Selection{{Book: 1, Set: 3}}}

	if !full.ContainsBook(1) || !full.ContainsBook(40) {
		t.Error("full scope should contain all books")
	}
	if !book.ContainsBook(1) || !book.ContainsBook(5) {
		t.Error("book scope should contain specified books")
	}
	if book.ContainsBook(2) {
		t.Error("book scope should not contain book 2")
	}
	if !set.ContainsBook(1) {
		t.Error("set scope should contain its book")
	}
	if set.ContainsBook(2) {
		t.Error("set scope should not contain book 2")
	}
}

func TestContainsSet(t *testing.T) {
	full := scope.Full()
	book := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}}}
	set := scope.Scope{Level: scope.LevelSet, Selections: []scope.Selection{{Book: 1, Set: 3}}}

	if !full.ContainsSet(1, 3) {
		t.Error("full scope should contain all sets")
	}
	if !book.ContainsSet(1, 3) || !book.ContainsSet(1, 10) {
		t.Error("book scope should contain all sets in its book")
	}
	if book.ContainsSet(2, 1) {
		t.Error("book scope should not contain sets in other books")
	}
	if !set.ContainsSet(1, 3) {
		t.Error("set scope should contain its set")
	}
	if set.ContainsSet(1, 4) {
		t.Error("set scope should not contain other sets")
	}
}

func TestExceeds(t *testing.T) {
	full := scope.Full()
	book1 := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}}}
	book15 := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}, {Book: 5}}}
	set13 := scope.Scope{Level: scope.LevelSet, Selections: []scope.Selection{{Book: 1, Set: 3}}}
	zero := scope.Scope{}

	// Full exceeds everything.
	if !full.Exceeds(book1) {
		t.Error("full should exceed book scope")
	}
	if !full.Exceeds(set13) {
		t.Error("full should exceed set scope")
	}
	if full.Exceeds(full) {
		t.Error("full should not exceed full")
	}

	// Book exceeds narrower book or set.
	if !book15.Exceeds(book1) {
		t.Error("B1,5 should exceed B1")
	}
	if book1.Exceeds(book15) {
		t.Error("B1 should not exceed B1,5")
	}
	if !book1.Exceeds(set13) {
		t.Error("book scope should exceed set scope")
	}

	// Exceeds zero (legacy) — anything clearing exceeds legacy.
	if !full.Exceeds(zero) {
		t.Error("full should exceed legacy (zero)")
	}
	if !book1.Exceeds(zero) {
		t.Error("book should exceed legacy (zero)")
	}
}

func TestBooksInScope(t *testing.T) {
	full := scope.Full()
	books := full.BooksInScope(40)
	if len(books) != 40 {
		t.Errorf("full scope BooksInScope: got %d, want 40", len(books))
	}

	s := scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 3}, {Book: 1}, {Book: 7}}}
	books = s.BooksInScope(40)
	if len(books) != 3 || books[0] != 1 || books[1] != 3 || books[2] != 7 {
		t.Errorf("book scope BooksInScope: got %v, want [1 3 7]", books)
	}
}

func TestParseSelectorsRangeBoundaries(t *testing.T) {
	tests := []struct {
		flag string
		want []scope.Selection
	}{
		// B40 — maximum valid book index
		{"B40", []scope.Selection{{Book: 40}}},
		// B1-40 — full book range as book scope
		{"B1-40", func() []scope.Selection {
			s := make([]scope.Selection, 40)
			for i := range s {
				s[i] = scope.Selection{Book: i + 1}
			}
			return s
		}()},
		// B1S10 — maximum set index
		{"B1S10", []scope.Selection{{Book: 1, Set: 10}}},
		// B1S3C0 — key 0 is valid
		{"B1S3C0", []scope.Selection{{Book: 1, Set: 3, Type: scope.TypeCtrl, Key: keyPtr(0)}}},
		// B1S3A9 — key 9 is valid
		{"B1S3A9", []scope.Selection{{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(9)}}},
	}
	for _, tt := range tests {
		s, err := scope.ParseSelectors([]string{tt.flag})
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", tt.flag, err)
		}
		if !selectionsEqual(s.Selections, tt.want) {
			t.Errorf("%q: selections = %v, want %v", tt.flag, s.Selections, tt.want)
		}
	}
}

func TestParseSelectorsRangeErrors(t *testing.T) {
	bad := []string{
		"B0",    // 0 < min (1)
		"B41",   // 41 > max (40)
		"B5-1",  // range end < start
		"B1S0",  // set 0 < min (1)
		"B1S11", // set 11 > max (10)
	}
	for _, flag := range bad {
		_, err := scope.ParseSelectors([]string{flag})
		if err == nil {
			t.Errorf("expected error for %q, got nil", flag)
		}
	}
}

func TestParseSelectorsBookCommaBook(t *testing.T) {
	// B1,B2 should work the same as B1,2.
	s1, err := scope.ParseSelectors([]string{"B1,B2"})
	if err != nil {
		t.Fatalf("B1,B2: %v", err)
	}
	s2, err := scope.ParseSelectors([]string{"B1,2"})
	if err != nil {
		t.Fatalf("B1,2: %v", err)
	}
	if s1.Level != s2.Level || !selectionsEqual(s1.Selections, s2.Selections) {
		t.Errorf("B1,B2 = %v, B1,2 = %v, expected equal", s1.Selections, s2.Selections)
	}
}

func TestParseSelectorsSetCommaSet(t *testing.T) {
	// B1S3,S4 — sibling set via tag prefix.
	s, err := scope.ParseSelectors([]string{"B1S3,S4"})
	if err != nil {
		t.Fatalf("B1S3,S4: %v", err)
	}
	if s.Level != scope.LevelSet {
		t.Errorf("level = %q, want set", s.Level)
	}
	want := []scope.Selection{{Book: 1, Set: 3}, {Book: 1, Set: 4}}
	if !selectionsEqual(s.Selections, want) {
		t.Errorf("selections = %v, want %v", s.Selections, want)
	}
}

func TestParseSelectorsZeroScope(t *testing.T) {
	// nil flags → full scope (identical to []string{} result).
	s, err := scope.ParseSelectors(nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != scope.LevelFull {
		t.Errorf("nil → level %q, want full", s.Level)
	}
}

func TestContainsMacroFullScope(t *testing.T) {
	full := scope.Full()
	// Full scope contains every macro slot.
	if !full.ContainsMacro(1, 1, scope.TypeCtrl, 0) {
		t.Error("full scope should contain ctrl 0")
	}
	if !full.ContainsMacro(40, 10, scope.TypeAlt, 9) {
		t.Error("full scope should contain alt 9")
	}
}

func TestContainsMacroWildKey(t *testing.T) {
	// A macro-level scope with nil Key (from C* or A*) must match every key of that type.
	s := scope.Scope{
		Level: scope.LevelMacro,
		Selections: []scope.Selection{
			{Book: 1, Set: 3, Type: scope.TypeCtrl}, // Key == nil → all ctrl
		},
	}
	for k := 0; k <= 9; k++ {
		if !s.ContainsMacro(1, 3, scope.TypeCtrl, k) {
			t.Errorf("wildcard ctrl should contain key %d", k)
		}
	}
	// Must not match alt keys.
	if s.ContainsMacro(1, 3, scope.TypeAlt, 0) {
		t.Error("wildcard ctrl should not match alt")
	}
	// Must not match a different set.
	if s.ContainsMacro(1, 4, scope.TypeCtrl, 0) {
		t.Error("wildcard ctrl should not match set 4")
	}
}

func TestParseSelectorsCtrlWildcardSibling(t *testing.T) {
	// B1S3A1,C* = A1 plus all ctrl keys; C* as a sibling must collapse to one selection.
	s, err := scope.ParseSelectors([]string{"B1S3A1,C*"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Level != scope.LevelMacro {
		t.Errorf("level = %q, want macro", s.Level)
	}
	want := []scope.Selection{
		{Book: 1, Set: 3, Type: scope.TypeAlt, Key: keyPtr(1)},
		{Book: 1, Set: 3, Type: scope.TypeCtrl}, // wildcard, no key
	}
	if !selectionsEqual(s.Selections, want) {
		t.Errorf("selections = %v, want %v", s.Selections, want)
	}
}

func keyPtr(n int) *int { return &n }

func selEqual(a, b scope.Selection) bool {
	if a.Book != b.Book || a.Set != b.Set || a.Type != b.Type {
		return false
	}
	if a.Key == nil && b.Key == nil {
		return true
	}
	if a.Key == nil || b.Key == nil {
		return false
	}
	return *a.Key == *b.Key
}

func selectionsEqual(a, b []scope.Selection) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !selEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}
