package scope_test

import (
	"strings"
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

// ---------------------------------------------------------------------------
// Comprehensive syntax table — one place for all valid parse cases.
// ---------------------------------------------------------------------------

func TestParseSelectorsSyntaxTable(t *testing.T) {
	k := keyPtr
	type tc struct {
		flags []string
		level scope.Level
		sels  []scope.Selection // nil means expect empty (full scope)
	}
	wc := func(book, set int, typ scope.MacroType) scope.Selection {
		return scope.Selection{Book: book, Set: set, Type: typ}
	}
	mk := func(book, set int, typ scope.MacroType, key int) scope.Selection {
		return scope.Selection{Book: book, Set: set, Type: typ, Key: k(key)}
	}
	tests := []tc{
		// --- Full scope ---
		{nil, scope.LevelFull, nil},
		{[]string{"*"}, scope.LevelFull, nil},
		{[]string{"B*"}, scope.LevelFull, nil},
		{[]string{"b*"}, scope.LevelFull, nil},
		{[]string{"B1,B*"}, scope.LevelFull, nil}, // B* anywhere → full

		// --- Book scope ---
		{[]string{"B1"}, scope.LevelBook, []scope.Selection{{Book: 1}}},
		{[]string{"b1"}, scope.LevelBook, []scope.Selection{{Book: 1}}}, // lowercase
		{[]string{"B30"}, scope.LevelBook, []scope.Selection{{Book: 30}}},
		{[]string{"B40"}, scope.LevelBook, []scope.Selection{{Book: 40}}}, // max
		{[]string{"B1,B30"}, scope.LevelBook, []scope.Selection{{Book: 1}, {Book: 30}}},
		{[]string{"B1,40"}, scope.LevelBook, []scope.Selection{{Book: 1}, {Book: 40}}},
		{[]string{"B1-3"}, scope.LevelBook, []scope.Selection{{Book: 1}, {Book: 2}, {Book: 3}}},
		{[]string{"B38-40"}, scope.LevelBook, []scope.Selection{{Book: 38}, {Book: 39}, {Book: 40}}},
		{[]string{"B1", "B40"}, scope.LevelBook, []scope.Selection{{Book: 1}, {Book: 40}}}, // multi-flag
		{[]string{"B1S*"}, scope.LevelBook, []scope.Selection{{Book: 1}}},                  // S* collapses
		{[]string{"B40S*"}, scope.LevelBook, []scope.Selection{{Book: 40}}},
		// single-element range (B1-1 = B1)
		{[]string{"B1-1"}, scope.LevelBook, []scope.Selection{{Book: 1}}},

		// --- Set scope ---
		{[]string{"B1S1"}, scope.LevelSet, []scope.Selection{{Book: 1, Set: 1}}},
		{[]string{"B30S5"}, scope.LevelSet, []scope.Selection{{Book: 30, Set: 5}}},
		{[]string{"B40S10"}, scope.LevelSet, []scope.Selection{{Book: 40, Set: 10}}}, // max
		{[]string{"b1s5"}, scope.LevelSet, []scope.Selection{{Book: 1, Set: 5}}},    // lowercase
		{[]string{"B1S1-3"}, scope.LevelSet, []scope.Selection{
			{Book: 1, Set: 1}, {Book: 1, Set: 2}, {Book: 1, Set: 3},
		}},
		{[]string{"B1S8-10"}, scope.LevelSet, []scope.Selection{
			{Book: 1, Set: 8}, {Book: 1, Set: 9}, {Book: 1, Set: 10},
		}},
		{[]string{"B1S1,5"}, scope.LevelSet, []scope.Selection{
			{Book: 1, Set: 1}, {Book: 1, Set: 5},
		}},
		{[]string{"B1S1,S10"}, scope.LevelSet, []scope.Selection{
			{Book: 1, Set: 1}, {Book: 1, Set: 10},
		}},
		{[]string{"B1S1", "B40S10"}, scope.LevelSet, []scope.Selection{
			{Book: 1, Set: 1}, {Book: 40, Set: 10},
		}},
		{[]string{"B1S1-1"}, scope.LevelSet, []scope.Selection{{Book: 1, Set: 1}}}, // single-element range

		// --- Macro scope: specific keys ---
		{[]string{"B1S1C0"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 0)}},
		{[]string{"B1S1C9"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 9)}},
		{[]string{"B1S1A0"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeAlt, 0)}},
		{[]string{"B1S1A9"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeAlt, 9)}},
		{[]string{"B30S5C3"}, scope.LevelMacro, []scope.Selection{mk(30, 5, scope.TypeCtrl, 3)}},
		{[]string{"B40S10A9"}, scope.LevelMacro, []scope.Selection{mk(40, 10, scope.TypeAlt, 9)}},
		{[]string{"b1s1c0"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 0)}}, // lowercase
		// Key ranges
		{[]string{"B1S1C0-9"}, scope.LevelMacro, func() []scope.Selection {
			s := make([]scope.Selection, 10)
			for i := range s {
				s[i] = mk(1, 1, scope.TypeCtrl, i)
			}
			return s
		}()},
		{[]string{"B1S1A0-9"}, scope.LevelMacro, func() []scope.Selection {
			s := make([]scope.Selection, 10)
			for i := range s {
				s[i] = mk(1, 1, scope.TypeAlt, i)
			}
			return s
		}()},
		{[]string{"B1S1C5-5"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 5)}}, // single-element range
		// Comma-separated key siblings
		{[]string{"B1S1A1,3"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeAlt, 1), mk(1, 1, scope.TypeAlt, 3)}},
		{[]string{"B1S1C2,A5"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 2), mk(1, 1, scope.TypeAlt, 5)}},

		// --- Macro scope: wildcards (nil key) ---
		{[]string{"B1S1C*"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl)}},
		{[]string{"B1S1A*"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeAlt)}},
		{[]string{"B40S10C*"}, scope.LevelMacro, []scope.Selection{wc(40, 10, scope.TypeCtrl)}},
		{[]string{"B40S10A*"}, scope.LevelMacro, []scope.Selection{wc(40, 10, scope.TypeAlt)}},
		// Implicit wildcard (bare C/A)
		{[]string{"B1S1C"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl)}},
		{[]string{"B1S1A"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeAlt)}},
		// Both types
		{[]string{"B1S1C,A"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl), wc(1, 1, scope.TypeAlt)}},
		{[]string{"B1S1A,C"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeAlt), wc(1, 1, scope.TypeCtrl)}},
		{[]string{"B1S1C*,A*"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl), wc(1, 1, scope.TypeAlt)}},
		{[]string{"B1S1C*,A"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl), wc(1, 1, scope.TypeAlt)}},
		{[]string{"B1S1C,A*"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl), wc(1, 1, scope.TypeAlt)}},
		{[]string{"B40S10C,A"}, scope.LevelMacro, []scope.Selection{wc(40, 10, scope.TypeCtrl), wc(40, 10, scope.TypeAlt)}},
		// Mixed: specific key + wildcard
		{[]string{"B1S1A1,C*"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeAlt, 1), wc(1, 1, scope.TypeCtrl)}},
		{[]string{"B1S1C5,A"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 5), wc(1, 1, scope.TypeAlt)}},
		// Multi-flag macro
		{[]string{"B1S1C*", "B40S10A*"}, scope.LevelMacro, []scope.Selection{wc(1, 1, scope.TypeCtrl), wc(40, 10, scope.TypeAlt)}},
		{[]string{"B1S1C0", "B40S10A9"}, scope.LevelMacro, []scope.Selection{mk(1, 1, scope.TypeCtrl, 0), mk(40, 10, scope.TypeAlt, 9)}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.flags, "+"), func(t *testing.T) {
			s, err := scope.ParseSelectors(tt.flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if s.Level != tt.level {
				t.Errorf("level = %q, want %q", s.Level, tt.level)
			}
			if tt.sels == nil {
				if len(s.Selections) != 0 {
					t.Errorf("selections = %v, want empty", s.Selections)
				}
			} else if !selectionsEqual(s.Selections, tt.sels) {
				t.Errorf("selections = %v\n          want %v", s.Selections, tt.sels)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Comprehensive error table — covers invalids, range violations, mixed levels,
// and security-oriented pathological inputs.
// ---------------------------------------------------------------------------

func TestParseSelectorsSyntaxErrors(t *testing.T) {
	tests := []struct {
		flags []string
		desc  string
	}{
		// Out-of-range values
		{[]string{"B0"}, "book 0 (below min 1)"},
		{[]string{"B41"}, "book 41 (above max 40)"},
		{[]string{"B1S0"}, "set 0 (below min 1)"},
		{[]string{"B1S11"}, "set 11 (above max 10)"},
		{[]string{"B1S1C10"}, "ctrl key 10 (above max 9)"},
		{[]string{"B1S1A10"}, "alt key 10 (above max 9)"},
		// Range boundary violations
		{[]string{"B5-1"}, "book range end < start"},
		{[]string{"B0-5"}, "book range start below min"},
		{[]string{"B1-41"}, "book range end above max"},
		{[]string{"B1S5-1"}, "set range end < start"},
		{[]string{"B1S0-5"}, "set range start below min"},
		{[]string{"B1S1-11"}, "set range end above max"},
		{[]string{"B1S1C5-3"}, "key range end < start"},
		{[]string{"B1S1C0-10"}, "key range end above max"},
		// Missing required context
		{[]string{"S1"}, "set without book"},
		{[]string{"C1"}, "ctrl without set"},
		{[]string{"A1"}, "alt without set"},
		// Mixed levels across flags
		{[]string{"B1", "B1S1"}, "book + set in separate flags"},
		{[]string{"B1S1", "B1S1C1"}, "set + macro in separate flags"},
		{[]string{"B1", "B1S1C1"}, "book + macro in separate flags"},
		// Mixed levels within a flag
		{[]string{"B1S3C1,S4"}, "macro then set sibling"},
		// Multiple books/sets before deeper component
		{[]string{"B1,2S3"}, "multiple comma-books before S"},
		{[]string{"B1-3S5"}, "range of books before S"},
		{[]string{"B1S2,3C4"}, "multiple comma-sets before C"},
		{[]string{"B1S2,3A4"}, "multiple comma-sets before A"},
		// Unknown / unexpected characters
		{[]string{"B1X"}, "unknown char X after number"},
		{[]string{"B1S1X"}, "unknown char X after set"},
		{[]string{"B1S1C1X"}, "trailing garbage after key"},
		{[]string{"B1-X"}, "non-digit range end"},
		// Bare type letters missing numspec illegally
		{[]string{"B1S1C,3"}, "bare C before digit sibling"},
		{[]string{"B1S1A,3"}, "bare A before digit sibling"},
		{[]string{"B1S1C,1,A"}, "bare C before digit even with A later"},
		// Empty / whitespace-only
		{[]string{""}, "empty string"},
		{[]string{" "}, "whitespace only"},
		{[]string{","}, "comma only"},
		// Bare letters with no numspec
		{[]string{"B"}, "bare B no numspec"},
		{[]string{"B1S"}, "bare S no numspec"},
		// ---------------------------------------------------------------
		// Security: unusual byte sequences that a parser might mishandle.
		// ---------------------------------------------------------------
		// Null bytes
		{[]string{"\x00"}, "NUL byte only"},
		{[]string{"B1\x00S1"}, "NUL byte mid-string"},
		{[]string{"B1S1C\x00"}, "NUL byte after C"},
		// ASCII control characters
		{[]string{"B1\nS1"}, "embedded newline (LF)"},
		{[]string{"B1\rS1"}, "embedded carriage return (CR)"},
		{[]string{"B1\tS1"}, "embedded tab"},
		{[]string{"B1\x01S1"}, "SOH control character"},
		// Injection-style separators
		{[]string{"B1;S1C1"}, "semicolon (SQL-style)"},
		{[]string{"B1|S1"}, "pipe character"},
		{[]string{"B1S1C1}B2"}, "closing brace injection"},
		{[]string{"B1S1C1<B2"}, "angle bracket injection"},
		// Very long numbers (integer overflow attempt)
		{[]string{"B" + strings.Repeat("9", 300)}, "huge book number (overflow)"},
		{[]string{"B1S" + strings.Repeat("9", 300)}, "huge set number (overflow)"},
		{[]string{"B1S1C" + strings.Repeat("9", 300)}, "huge key number (overflow)"},
		// Negative number attempts (hyphen at start of numspec)
		{[]string{"B-1"}, "leading hyphen on book"},
		{[]string{"B1S-1"}, "leading hyphen on set"},
		{[]string{"B1S1C-5"}, "leading hyphen on key"},
		// Double special characters
		{[]string{"B1S1C**"}, "double wildcard after C (second * is trailing garbage)"},
		{[]string{"B1S1A**"}, "double wildcard after A"},
		{[]string{"B1--3"}, "double dash in book range"},
		{[]string{"B1S1C--5"}, "double dash in key range"},
		// Unicode that superficially resembles ASCII letters
		{[]string{"B1S1Α"}, "U+0391 Greek capital alpha (looks like A)"},
		{[]string{"B1S1Ⓒ"}, "U+24B8 circled C"},
		// Multi-byte UTF-8 sequences
		{[]string{"B1S1\xc3\xa9"}, "multi-byte UTF-8 (é = 0xC3 0xA9)"},
		{[]string{"B1S1\xc0\x80"}, "overlong UTF-8 NUL (0xC0 0x80)"},
		{[]string{"B1S1\xff\xfe"}, "invalid UTF-8 sequence"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := scope.ParseSelectors(tt.flags)
			if err == nil {
				t.Errorf("flag %q: expected error, got nil", tt.flags)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ContainsMacro / ContainsSet / ContainsBook comprehensive table.
// ---------------------------------------------------------------------------

func TestContainsMacroTable(t *testing.T) {
	type macroQ struct {
		book, set int
		typ       scope.MacroType
		key       int
	}
	type tc struct {
		name        string
		sc          scope.Scope
		contains    []macroQ
		notContains []macroQ
	}
	k := keyPtr

	tests := []tc{
		{
			name: "full scope contains every slot",
			sc:   scope.Full(),
			contains: []macroQ{
				{1, 1, scope.TypeCtrl, 0},
				{1, 1, scope.TypeCtrl, 9},
				{1, 1, scope.TypeAlt, 0},
				{40, 10, scope.TypeAlt, 9},
				{30, 5, scope.TypeCtrl, 5},
			},
		},
		{
			name: "book scope contains all macros in book, none outside",
			sc:   scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 1}}},
			contains: []macroQ{
				{1, 1, scope.TypeCtrl, 0},
				{1, 10, scope.TypeAlt, 9},
				{1, 5, scope.TypeCtrl, 5},
			},
			notContains: []macroQ{
				{2, 1, scope.TypeCtrl, 0},
				{40, 5, scope.TypeAlt, 0},
			},
		},
		{
			name: "book 40 scope",
			sc:   scope.Scope{Level: scope.LevelBook, Selections: []scope.Selection{{Book: 40}}},
			contains: []macroQ{
				{40, 10, scope.TypeAlt, 9},
			},
			notContains: []macroQ{
				{39, 10, scope.TypeAlt, 9},
				{1, 1, scope.TypeCtrl, 0},
			},
		},
		{
			name: "set scope contains all macros in set, none in adjacent set/book",
			sc:   scope.Scope{Level: scope.LevelSet, Selections: []scope.Selection{{Book: 1, Set: 5}}},
			contains: []macroQ{
				{1, 5, scope.TypeCtrl, 0},
				{1, 5, scope.TypeCtrl, 9},
				{1, 5, scope.TypeAlt, 0},
				{1, 5, scope.TypeAlt, 9},
			},
			notContains: []macroQ{
				{1, 4, scope.TypeCtrl, 0},
				{1, 6, scope.TypeAlt, 0},
				{2, 5, scope.TypeCtrl, 0},
			},
		},
		{
			name: "ctrl wildcard (nil key) matches all ctrl keys, not alt",
			sc: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{
				{Book: 1, Set: 1, Type: scope.TypeCtrl},
			}},
			contains: []macroQ{
				{1, 1, scope.TypeCtrl, 0},
				{1, 1, scope.TypeCtrl, 5},
				{1, 1, scope.TypeCtrl, 9},
			},
			notContains: []macroQ{
				{1, 1, scope.TypeAlt, 0},
				{1, 1, scope.TypeAlt, 9},
				{2, 1, scope.TypeCtrl, 0},
				{1, 2, scope.TypeCtrl, 0},
			},
		},
		{
			name: "alt wildcard (nil key) matches all alt keys, not ctrl",
			sc: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{
				{Book: 1, Set: 1, Type: scope.TypeAlt},
			}},
			contains: []macroQ{
				{1, 1, scope.TypeAlt, 0},
				{1, 1, scope.TypeAlt, 9},
			},
			notContains: []macroQ{
				{1, 1, scope.TypeCtrl, 0},
				{2, 1, scope.TypeAlt, 0},
			},
		},
		{
			name: "specific key 0 is distinct from wildcard — only matches key 0",
			sc: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{
				{Book: 1, Set: 1, Type: scope.TypeCtrl, Key: k(0)},
			}},
			contains:    []macroQ{{1, 1, scope.TypeCtrl, 0}},
			notContains: []macroQ{{1, 1, scope.TypeCtrl, 1}, {1, 1, scope.TypeCtrl, 9}, {1, 1, scope.TypeAlt, 0}},
		},
		{
			name: "specific key 9 only matches key 9",
			sc: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{
				{Book: 40, Set: 10, Type: scope.TypeAlt, Key: k(9)},
			}},
			contains:    []macroQ{{40, 10, scope.TypeAlt, 9}},
			notContains: []macroQ{{40, 10, scope.TypeAlt, 8}, {40, 10, scope.TypeCtrl, 9}, {1, 1, scope.TypeAlt, 9}},
		},
		{
			name: "both ctrl and alt wildcards cover both types",
			sc: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{
				{Book: 1, Set: 1, Type: scope.TypeCtrl},
				{Book: 1, Set: 1, Type: scope.TypeAlt},
			}},
			contains: []macroQ{
				{1, 1, scope.TypeCtrl, 0},
				{1, 1, scope.TypeCtrl, 9},
				{1, 1, scope.TypeAlt, 0},
				{1, 1, scope.TypeAlt, 9},
			},
			notContains: []macroQ{
				{2, 1, scope.TypeCtrl, 0},
				{1, 2, scope.TypeAlt, 0},
			},
		},
		{
			name: "multi-book multi-set macro scope",
			sc: scope.Scope{Level: scope.LevelMacro, Selections: []scope.Selection{
				{Book: 1, Set: 1, Type: scope.TypeCtrl, Key: k(5)},
				{Book: 40, Set: 10, Type: scope.TypeAlt, Key: k(9)},
			}},
			contains:    []macroQ{{1, 1, scope.TypeCtrl, 5}, {40, 10, scope.TypeAlt, 9}},
			notContains: []macroQ{{1, 1, scope.TypeCtrl, 4}, {40, 10, scope.TypeAlt, 8}, {2, 1, scope.TypeCtrl, 5}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, q := range tt.contains {
				if !tt.sc.ContainsMacro(q.book, q.set, q.typ, q.key) {
					t.Errorf("should contain {book:%d, set:%d, type:%s, key:%d}", q.book, q.set, q.typ, q.key)
				}
			}
			for _, q := range tt.notContains {
				if tt.sc.ContainsMacro(q.book, q.set, q.typ, q.key) {
					t.Errorf("should NOT contain {book:%d, set:%d, type:%s, key:%d}", q.book, q.set, q.typ, q.key)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Quirky-but-valid inputs: document parser edge cases explicitly.
// ---------------------------------------------------------------------------

func TestParseSelectorsQuirks(t *testing.T) {
	// Trailing comma is silently consumed; B1, = B1.
	s, err := scope.ParseSelectors([]string{"B1,"})
	if err != nil {
		t.Errorf("B1, (trailing comma): unexpected error: %v", err)
	} else if s.Level != scope.LevelBook || len(s.Selections) != 1 || s.Selections[0].Book != 1 {
		t.Errorf("B1, = %v, want [{Book:1}]", s.Selections)
	}

	// Leading commas are silently consumed; ,B1 = B1.
	s, err = scope.ParseSelectors([]string{",B1"})
	if err != nil {
		t.Errorf(",B1 (leading comma): unexpected error: %v", err)
	} else if s.Level != scope.LevelBook || len(s.Selections) != 1 || s.Selections[0].Book != 1 {
		t.Errorf(",B1 = %v, want [{Book:1}]", s.Selections)
	}

	// B** = full scope (first * triggers full scope, second * never reached).
	s, err = scope.ParseSelectors([]string{"B**"})
	if err != nil {
		t.Errorf("B** (double wildcard at book level): unexpected error: %v", err)
	} else if s.Level != scope.LevelFull {
		t.Errorf("B** level = %q, want full", s.Level)
	}

	// Mixed-case is normalised to upper.
	s, err = scope.ParseSelectors([]string{"b40s10a9"})
	if err != nil {
		t.Errorf("b40s10a9 (fully lowercase): unexpected error: %v", err)
	} else if s.Level != scope.LevelMacro {
		t.Errorf("b40s10a9 level = %q, want macro", s.Level)
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
