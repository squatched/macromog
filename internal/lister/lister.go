package lister

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/squatched/macromog/internal/dat"
)

// CharacterInfo summarizes one FFXI character directory.
type CharacterInfo struct {
	ID        string
	Dir       string
	BookCount int
}

// BookInfo describes one populated macro book within a character directory.
type BookInfo struct {
	Index    int
	Name     string
	SetCount int
}

// DiscoverCharacters scans userDir for FFXI character subdirectories and
// returns a summary of each, sorted by directory name.
func DiscoverCharacters(userDir string) ([]CharacterInfo, error) {
	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil, err
	}

	var chars []CharacterInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(userDir, e.Name())
		if !IsCharacterDir(dir) {
			continue
		}
		books, err := BooksForCharacter(dir)
		if err != nil {
			continue
		}
		chars = append(chars, CharacterInfo{
			ID:        e.Name(),
			Dir:       dir,
			BookCount: len(books),
		})
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i].ID < chars[j].ID })
	return chars, nil
}

// IsCharacterDir reports whether dir contains mcr.dat, the sentinel for an
// FFXI character USER directory.
func IsCharacterDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "mcr.dat"))
	return err == nil
}

// BooksForCharacter returns non-empty book summaries for charDir, sorted by
// book index. Only books with at least one non-empty macro set are included.
func BooksForCharacter(charDir string) ([]BookInfo, error) {
	titles, err := dat.ReadBookTitles(charDir)
	if err != nil {
		return nil, err
	}

	files, err := dat.DiscoverMacroFiles(charDir)
	if err != nil {
		return nil, err
	}

	type bookAccum struct {
		name     string
		setCount int
	}
	books := make(map[int]*bookAccum)

	for _, path := range files {
		index, ok := dat.ParseMacroFileName(filepath.Base(path))
		if !ok {
			continue
		}
		book, _ := dat.ParseFileIndex(index)
		if book < 1 || book > dat.MaxBooks {
			continue
		}

		ms, err := dat.ReadMacroSetFile(path)
		if err != nil {
			continue
		}

		if !macroSetHasContent(ms) {
			continue
		}

		if _, ok := books[book]; !ok {
			books[book] = &bookAccum{name: titles[book-1]}
		}
		books[book].setCount++
	}

	result := make([]BookInfo, 0, len(books))
	for bk, acc := range books {
		result = append(result, BookInfo{Index: bk, Name: acc.name, SetCount: acc.setCount})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Index < result[j].Index })
	return result, nil
}

func macroSetHasContent(ms dat.MacroSet) bool {
	for i := 0; i < 10; i++ {
		if !ms.Ctrl[i].Empty() || !ms.Alt[i].Empty() {
			return true
		}
	}
	return false
}
