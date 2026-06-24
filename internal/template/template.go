package template

import (
	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
)

// Generate returns a blank Document pre-structured for the given scope.
// Every macro slot within scope is present with an empty name and six empty
// content lines. The exported_at field is omitted (templates are not exports).
// Character is included only when non-empty.
func Generate(sc scope.Scope, character string) export.Document {
	if sc.IsZero() {
		sc = scope.Full()
	}

	doc := export.Document{
		Version:   1,
		Character: character,
		Scope:     sc,
		Books:     make(map[int]export.Book),
	}

	switch sc.Level {
	case scope.LevelFull:
		for book := 1; book <= dat.MaxBooks; book++ {
			doc.Books[book] = templateBook(book, allSets(), sc)
		}

	case scope.LevelBook:
		for _, sel := range sc.Selections {
			doc.Books[sel.Book] = templateBook(sel.Book, allSets(), sc)
		}

	case scope.LevelSet:
		// Group selections by book.
		bookSets := make(map[int][]int)
		for _, sel := range sc.Selections {
			bookSets[sel.Book] = append(bookSets[sel.Book], sel.Set)
		}
		for bookIdx, sets := range bookSets {
			doc.Books[bookIdx] = templateBook(bookIdx, sets, sc)
		}

	case scope.LevelMacro:
		// Group by book → sets.
		type setKey struct{ book, set int }
		bookSets := make(map[int]map[int]bool)
		for _, sel := range sc.Selections {
			if bookSets[sel.Book] == nil {
				bookSets[sel.Book] = make(map[int]bool)
			}
			bookSets[sel.Book][sel.Set] = true
		}
		for bookIdx, setsMap := range bookSets {
			sets := make([]int, 0, len(setsMap))
			for s := range setsMap {
				sets = append(sets, s)
			}
			doc.Books[bookIdx] = templateBook(bookIdx, sets, sc)
		}
	}

	return doc
}

func allSets() []int {
	sets := make([]int, dat.SetsPerBook)
	for i := range sets {
		sets[i] = i + 1
	}
	return sets
}

func templateBook(bookIdx int, sets []int, sc scope.Scope) export.Book {
	b := export.Book{
		Name: "",
		Sets: make(map[int]export.Set),
	}
	for _, setIdx := range sets {
		b.Sets[setIdx] = templateSet(bookIdx, setIdx, sc)
	}
	return b
}

func templateSet(bookIdx, setIdx int, sc scope.Scope) export.Set {
	s := export.Set{}
	for i := 0; i < dat.SetsPerBook; i++ {
		yamlKey := dat.YAMLKey(i)
		if sc.ContainsMacro(bookIdx, setIdx, scope.TypeCtrl, yamlKey) {
			if s.Ctrl == nil {
				s.Ctrl = make(map[int]export.Macro)
			}
			s.Ctrl[yamlKey] = emptyMacro()
		}
		if sc.ContainsMacro(bookIdx, setIdx, scope.TypeAlt, yamlKey) {
			if s.Alt == nil {
				s.Alt = make(map[int]export.Macro)
			}
			s.Alt[yamlKey] = emptyMacro()
		}
	}
	return s
}

func emptyMacro() export.Macro {
	return export.Macro{
		Name:     "",
		Contents: make([]string, dat.LineCount),
	}
}
