package scope

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Level describes how wide a scope's authority is.
type Level string

const (
	LevelFull  Level = "full"
	LevelBook  Level = "book"
	LevelSet   Level = "set"
	LevelMacro Level = "macro"
)

// MacroType is "ctrl" or "alt".
type MacroType string

const (
	TypeCtrl MacroType = "ctrl"
	TypeAlt  MacroType = "alt"
)

// Selection is one entry in a scope's selections list.
// Fields present depend on the scope level:
//
//	book:  Book only
//	set:   Book, Set
//	macro: Book, Set, Type, Key
type Selection struct {
	Book int       `yaml:"book"`
	Set  int       `yaml:"set,omitempty"`
	Type MacroType `yaml:"type,omitempty"`
	Key  int       `yaml:"key,omitempty"`
}

// Scope is the authority boundary embedded in a YAML document.
type Scope struct {
	Level      Level       `yaml:"level"`
	Selections []Selection `yaml:"selections,omitempty"`
}

// IsZero reports whether s is an unset scope (not present in the YAML).
// This is distinct from LevelFull, which is an explicit full-scope declaration.
func (s Scope) IsZero() bool { return s.Level == "" }

// Full returns a full-scope Scope value.
func Full() Scope { return Scope{Level: LevelFull} }

// errFullScope is an internal sentinel returned by the parser when it detects
// a wildcard that implies full scope (e.g. B*).
var errFullScope = errors.New("full scope")

// ParseSelectors parses one or more --scope flag values into a Scope.
// Returns Full() when flags is empty.
func ParseSelectors(flags []string) (Scope, error) {
	if len(flags) == 0 {
		return Full(), nil
	}

	var allSels []Selection
	var inferredLevel Level

	for _, flag := range flags {
		sels, level, err := parseFlag(flag)
		if err != nil {
			return Scope{}, fmt.Errorf("--scope %q: %w", flag, err)
		}
		if level == LevelFull {
			return Full(), nil
		}
		if inferredLevel == "" {
			inferredLevel = level
		} else if inferredLevel != level {
			return Scope{}, fmt.Errorf(
				"--scope: mixed levels (%s and %s); all --scope flags must be at the same granularity",
				inferredLevel, level,
			)
		}
		allSels = append(allSels, sels...)
	}

	return Scope{Level: inferredLevel, Selections: allSels}, nil
}

// parseFlag parses a single --scope flag value (which may contain commas).
func parseFlag(flag string) ([]Selection, Level, error) {
	upper := strings.ToUpper(strings.TrimSpace(flag))
	if upper == "*" || upper == "B*" {
		return nil, LevelFull, nil
	}
	p := &parser{input: upper}
	sels, err := p.parse()
	if errors.Is(err, errFullScope) {
		return nil, LevelFull, nil
	}
	if err != nil {
		return nil, "", err
	}
	if len(sels) == 0 {
		return nil, "", fmt.Errorf("empty selector")
	}
	return sels, inferLevel(sels), nil
}

// inferLevel returns the scope Level implied by a slice of selections.
func inferLevel(sels []Selection) Level {
	for _, s := range sels {
		if s.Type != "" {
			return LevelMacro
		}
	}
	for _, s := range sels {
		if s.Set != 0 {
			return LevelSet
		}
	}
	if len(sels) > 0 {
		return LevelBook
	}
	return LevelFull
}

// parser is a hand-rolled recursive descent parser for scope selectors.
//
// Supported grammar:
//
//	selector   = book_seg ( ',' sibling )*
//	book_seg   = 'B' numspec [ set_seg ]
//	set_seg    = 'S' numspec [ key_seg ]
//	key_seg    = ( 'C' | 'A' ) numspec
//	numspec    = '*' | number [ '-' number ]
//	sibling    = book_seg | set_seg | key_seg | numspec   (inherits parent type)
//
// Wildcards at the book level ('B*') signal full scope.
// Wildcards at the set level ('B<n>S*') produce a book-level selection.
// Multiple books before S or multiple sets before C/A are disallowed; use
// separate --scope flags instead.
type parser struct {
	input string
	pos   int
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) atEnd() bool { return p.pos >= len(p.input) }

func (p *parser) consume() byte {
	b := p.input[p.pos]
	p.pos++
	return b
}

// pctx tracks inter-sibling context so commas can produce siblings.
type pctx struct {
	book    int
	set     int
	hasBook bool
	hasSet  bool
	lastTag byte // most recent component letter: B, S, C, A
	depth   byte // deepest emission level: B, S, M (macro) — zero until first emit
}

func (p *parser) parse() ([]Selection, error) {
	var c pctx
	var sels []Selection

	for !p.atEnd() {
		ch := p.peek()
		if ch == ',' {
			p.consume()
			continue
		}

		switch {
		// --- Book component ---
		case ch == 'B':
			p.consume()
			nums, wild, err := p.parseNumSpec(1, 40)
			if err != nil {
				return nil, fmt.Errorf("book: %w", err)
			}
			if wild {
				// B* = full scope regardless of what follows.
				return nil, errFullScope
			}
			if len(nums) > 1 && p.peek() == 'S' {
				return nil, fmt.Errorf("cannot specify multiple books before S; use separate --scope flags")
			}
			c.lastTag = 'B'
			if p.peek() == 'S' {
				// Book context only — set component will emit the selection.
				c.book = nums[0]
				c.hasBook = true
			} else {
				if c.depth != 0 && c.depth != 'B' {
					return nil, fmt.Errorf("cannot mix book-level selector with deeper-level selectors in one --scope value")
				}
				for _, n := range nums {
					sels = append(sels, Selection{Book: n})
				}
				c.depth = 'B'
				c.book = nums[len(nums)-1]
				c.hasBook = true
				c.hasSet = false
			}

		// --- Set component ---
		case ch == 'S':
			if !c.hasBook {
				return nil, fmt.Errorf("S (set) requires a preceding B (book) component")
			}
			p.consume()
			nums, wild, err := p.parseNumSpec(1, 10)
			if err != nil {
				return nil, fmt.Errorf("set: %w", err)
			}
			c.lastTag = 'S'
			if wild {
				// B<n>S* = all sets in that book = book-level authority.
				if c.depth != 0 && c.depth != 'B' {
					return nil, fmt.Errorf("cannot mix book-level selector with deeper-level selectors in one --scope value")
				}
				sels = append(sels, Selection{Book: c.book})
				c.depth = 'B'
				break
			}
			if len(nums) > 1 && (p.peek() == 'C' || p.peek() == 'A') {
				return nil, fmt.Errorf("cannot specify multiple sets before C/A; use separate --scope flags")
			}
			if p.peek() == 'C' || p.peek() == 'A' {
				// Set context only — key component will emit the selection.
				c.set = nums[0]
				c.hasSet = true
			} else {
				if c.depth != 0 && c.depth != 'S' {
					return nil, fmt.Errorf("cannot mix set-level selector with other-level selectors in one --scope value")
				}
				for _, n := range nums {
					sels = append(sels, Selection{Book: c.book, Set: n})
				}
				c.depth = 'S'
				c.set = nums[len(nums)-1]
				c.hasSet = true
			}

		// --- Ctrl key component ---
		case ch == 'C':
			if !c.hasSet {
				return nil, fmt.Errorf("C (ctrl) requires a preceding S (set) component")
			}
			p.consume()
			nums, wild, err := p.parseNumSpec(0, 9)
			if err != nil {
				return nil, fmt.Errorf("ctrl: %w", err)
			}
			c.lastTag = 'C'
			if c.depth != 0 && c.depth != 'M' {
				return nil, fmt.Errorf("cannot mix macro-level selector with other-level selectors in one --scope value")
			}
			c.depth = 'M'
			if wild {
				for k := 0; k <= 9; k++ {
					sels = append(sels, Selection{Book: c.book, Set: c.set, Type: TypeCtrl, Key: k})
				}
				break
			}
			for _, n := range nums {
				sels = append(sels, Selection{Book: c.book, Set: c.set, Type: TypeCtrl, Key: n})
			}

		// --- Alt key component ---
		case ch == 'A':
			if !c.hasSet {
				return nil, fmt.Errorf("A (alt) requires a preceding S (set) component")
			}
			p.consume()
			nums, wild, err := p.parseNumSpec(0, 9)
			if err != nil {
				return nil, fmt.Errorf("alt: %w", err)
			}
			c.lastTag = 'A'
			if c.depth != 0 && c.depth != 'M' {
				return nil, fmt.Errorf("cannot mix macro-level selector with other-level selectors in one --scope value")
			}
			c.depth = 'M'
			if wild {
				for k := 0; k <= 9; k++ {
					sels = append(sels, Selection{Book: c.book, Set: c.set, Type: TypeAlt, Key: k})
				}
				break
			}
			for _, n := range nums {
				sels = append(sels, Selection{Book: c.book, Set: c.set, Type: TypeAlt, Key: n})
			}

		// --- Bare number: sibling inheriting lastTag ---
		case ch >= '0' && ch <= '9':
			if c.lastTag == 0 {
				return nil, fmt.Errorf("bare number without a preceding component type (B/S/C/A)")
			}
			switch c.lastTag {
			case 'B':
				nums, _, err := p.parseNumSpec(1, 40)
				if err != nil {
					return nil, fmt.Errorf("book sibling: %w", err)
				}
				if c.depth != 0 && c.depth != 'B' {
					return nil, fmt.Errorf("cannot mix book-level selector with deeper-level selectors in one --scope value")
				}
				for _, n := range nums {
					sels = append(sels, Selection{Book: n})
				}
				c.depth = 'B'
				c.book = nums[len(nums)-1]
				c.hasSet = false
			case 'S':
				nums, _, err := p.parseNumSpec(1, 10)
				if err != nil {
					return nil, fmt.Errorf("set sibling: %w", err)
				}
				if c.depth != 0 && c.depth != 'S' {
					return nil, fmt.Errorf("cannot mix set-level selector with other-level selectors in one --scope value")
				}
				for _, n := range nums {
					sels = append(sels, Selection{Book: c.book, Set: n})
				}
				c.depth = 'S'
				c.set = nums[len(nums)-1]
				c.hasSet = true
			case 'C':
				nums, _, err := p.parseNumSpec(0, 9)
				if err != nil {
					return nil, fmt.Errorf("ctrl sibling: %w", err)
				}
				if c.depth != 0 && c.depth != 'M' {
					return nil, fmt.Errorf("cannot mix macro-level selector with other-level selectors in one --scope value")
				}
				c.depth = 'M'
				for _, n := range nums {
					sels = append(sels, Selection{Book: c.book, Set: c.set, Type: TypeCtrl, Key: n})
				}
			case 'A':
				nums, _, err := p.parseNumSpec(0, 9)
				if err != nil {
					return nil, fmt.Errorf("alt sibling: %w", err)
				}
				if c.depth != 0 && c.depth != 'M' {
					return nil, fmt.Errorf("cannot mix macro-level selector with other-level selectors in one --scope value")
				}
				c.depth = 'M'
				for _, n := range nums {
					sels = append(sels, Selection{Book: c.book, Set: c.set, Type: TypeAlt, Key: n})
				}
			}

		default:
			return nil, fmt.Errorf("unexpected character %q at position %d", string(ch), p.pos)
		}
	}

	if !p.atEnd() {
		return nil, fmt.Errorf("unexpected trailing input at position %d", p.pos)
	}
	return sels, nil
}

// parseNumSpec parses "*", a single integer, or a range "n-m".
// Returns the expanded list and whether the input was a wildcard.
func (p *parser) parseNumSpec(min, max int) (nums []int, wild bool, err error) {
	if p.peek() == '*' {
		p.consume()
		all := make([]int, 0, max-min+1)
		for i := min; i <= max; i++ {
			all = append(all, i)
		}
		return all, true, nil
	}

	first, err := p.parseNumber()
	if err != nil {
		return nil, false, err
	}
	if first < min || first > max {
		return nil, false, fmt.Errorf("value %d out of range %d–%d", first, min, max)
	}

	if p.peek() == '-' {
		p.consume()
		last, err := p.parseNumber()
		if err != nil {
			return nil, false, fmt.Errorf("range end: %w", err)
		}
		if last < first || last > max {
			return nil, false, fmt.Errorf("range end %d out of range %d–%d", last, first, max)
		}
		all := make([]int, 0, last-first+1)
		for i := first; i <= last; i++ {
			all = append(all, i)
		}
		return all, false, nil
	}

	return []int{first}, false, nil
}

func (p *parser) parseNumber() (int, error) {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if p.pos == start {
		if p.atEnd() {
			return 0, fmt.Errorf("expected number, got end of input")
		}
		return 0, fmt.Errorf("expected number, got %q", string(p.input[p.pos]))
	}
	return strconv.Atoi(p.input[start:p.pos])
}

// ContainsBook reports whether s has authority over the given 1-based book index.
func (s Scope) ContainsBook(book int) bool {
	if s.Level == LevelFull {
		return true
	}
	for _, sel := range s.Selections {
		if sel.Book == book {
			return true
		}
	}
	return false
}

// ContainsSet reports whether s has authority over the given book+set pair.
func (s Scope) ContainsSet(book, set int) bool {
	switch s.Level {
	case LevelFull:
		return true
	case LevelBook:
		return s.ContainsBook(book)
	case LevelSet, LevelMacro:
		for _, sel := range s.Selections {
			if sel.Book == book && sel.Set == set {
				return true
			}
		}
	}
	return false
}

// ContainsMacro reports whether s has authority over the given macro slot.
func (s Scope) ContainsMacro(book, set int, mtype MacroType, key int) bool {
	switch s.Level {
	case LevelFull:
		return true
	case LevelBook:
		return s.ContainsBook(book)
	case LevelSet:
		return s.ContainsSet(book, set)
	case LevelMacro:
		for _, sel := range s.Selections {
			if sel.Book == book && sel.Set == set && sel.Type == mtype && sel.Key == key {
				return true
			}
		}
	}
	return false
}

// Exceeds reports whether s claims authority that other does not cover.
// Used to determine if an import --scope override requires a confirmation prompt.
func (s Scope) Exceeds(other Scope) bool {
	if other.IsZero() {
		// Legacy YAML has no scope; any scope with clearing semantics exceeds it.
		return s.Level != LevelMacro
	}
	levelOrder := map[Level]int{
		LevelMacro: 0,
		LevelSet:   1,
		LevelBook:  2,
		LevelFull:  3,
	}
	if levelOrder[s.Level] > levelOrder[other.Level] {
		return true
	}
	if s.Level == other.Level && s.Level != LevelFull {
		for _, sel := range s.Selections {
			if !other.containsSelection(sel) {
				return true
			}
		}
	}
	return false
}

func (s Scope) containsSelection(sel Selection) bool {
	for _, existing := range s.Selections {
		if existing == sel {
			return true
		}
	}
	return false
}

// BooksInScope returns all book indices (1–maxBook) that fall within s,
// in ascending order.
func (s Scope) BooksInScope(maxBook int) []int {
	if s.Level == LevelFull {
		books := make([]int, maxBook)
		for i := range books {
			books[i] = i + 1
		}
		return books
	}
	seen := make(map[int]bool)
	for _, sel := range s.Selections {
		seen[sel.Book] = true
	}
	books := make([]int, 0, len(seen))
	for i := 1; i <= maxBook; i++ {
		if seen[i] {
			books = append(books, i)
		}
	}
	return books
}
