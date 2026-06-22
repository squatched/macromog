package lister_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/lister"
)

// blankDat returns a valid-header, all-zero macro set file (no macros).
func blankDat() []byte {
	buf := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(buf[0:4], dat.MagicVersion)
	return buf
}

// namedDat returns a dat file where ctrl[0] has the given macro name.
// The name must be at most 8 ASCII bytes.
func namedDat(name string) []byte {
	buf := blankDat()
	// ctrl[0] name offset: HeaderSize + MacroPrefixSize + LineCount*LineSize
	nameOff := dat.HeaderSize + dat.MacroPrefixSize + dat.LineCount*dat.LineSize
	copy(buf[nameOff:nameOff+dat.NameSize], []byte(name))
	return buf
}

// makeUserDir creates a temp USER directory with the specified character
// subdirectories. The charFiles map keys are character IDs; the value is a
// map of filename → file content to place inside each character directory.
func makeUserDir(t *testing.T, charFiles map[string]map[string][]byte) string {
	t.Helper()
	userDir := t.TempDir()
	for id, files := range charFiles {
		charDir := filepath.Join(userDir, id)
		if err := os.MkdirAll(charDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(charDir, name), content, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return userDir
}

func TestDiscoverCharacters_Empty(t *testing.T) {
	userDir := t.TempDir()
	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) != 0 {
		t.Errorf("expected 0 chars, got %d", len(chars))
	}
}

func TestDiscoverCharacters_NonCharDir(t *testing.T) {
	// A subdirectory without mcr.dat should not be recognized as a character dir.
	userDir := makeUserDir(t, map[string]map[string][]byte{
		"notachar": {"other.dat": []byte("irrelevant")},
	})
	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) != 0 {
		t.Errorf("expected 0 chars, got %d", len(chars))
	}
}

func TestDiscoverCharacters_SortedByID(t *testing.T) {
	userDir := makeUserDir(t, map[string]map[string][]byte{
		"zzzzzz": {"mcr.dat": blankDat()},
		"aaaaaa": {"mcr.dat": blankDat()},
		"mmmmmm": {"mcr.dat": blankDat()},
	})
	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) != 3 {
		t.Fatalf("expected 3 chars, got %d", len(chars))
	}
	if chars[0].ID != "aaaaaa" || chars[1].ID != "mmmmmm" || chars[2].ID != "zzzzzz" {
		t.Errorf("unexpected order: %v %v %v", chars[0].ID, chars[1].ID, chars[2].ID)
	}
}

func TestDiscoverCharacters_BookCount(t *testing.T) {
	// One char with two non-empty sets in book 1 (mcr.dat and mcr1.dat).
	// Another char with no macros.
	userDir := makeUserDir(t, map[string]map[string][]byte{
		"aabbcc": {
			"mcr.dat":  namedDat("Macro1"),
			"mcr1.dat": namedDat("Macro2"),
		},
		"ddeeff": {
			"mcr.dat": blankDat(),
		},
	})

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) != 2 {
		t.Fatalf("expected 2 chars, got %d", len(chars))
	}

	// aabbcc: mcr.dat and mcr1.dat are both book 1, so book count = 1.
	if chars[0].ID != "aabbcc" || chars[0].BookCount != 1 {
		t.Errorf("aabbcc: got BookCount=%d, want 1", chars[0].BookCount)
	}
	// ddeeff: blank dat, no macros.
	if chars[1].ID != "ddeeff" || chars[1].BookCount != 0 {
		t.Errorf("ddeeff: got BookCount=%d, want 0", chars[1].BookCount)
	}
}

func TestDiscoverCharacters_MissingDir(t *testing.T) {
	_, err := lister.DiscoverCharacters("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestBooksForCharacter_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), blankDat(), 0o644); err != nil {
		t.Fatal(err)
	}
	books, err := lister.BooksForCharacter(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 0 {
		t.Errorf("expected 0 books, got %d", len(books))
	}
}

func TestBooksForCharacter_SingleBook(t *testing.T) {
	dir := t.TempDir()
	// mcr.dat = book 1 set 1, mcr1.dat = book 1 set 2 — both non-empty.
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), namedDat("SetA"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mcr1.dat"), namedDat("SetB"), 0o644); err != nil {
		t.Fatal(err)
	}

	books, err := lister.BooksForCharacter(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(books))
	}
	if books[0].Index != 1 || books[0].SetCount != 2 {
		t.Errorf("book: index=%d setCount=%d, want index=1 setCount=2", books[0].Index, books[0].SetCount)
	}
}

func TestBooksForCharacter_MultipleBooks(t *testing.T) {
	dir := t.TempDir()
	// book 1 set 1 = mcr.dat, book 33 set 1 = mcr320.dat
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), namedDat("B1S1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mcr320.dat"), namedDat("B33S1"), 0o644); err != nil {
		t.Fatal(err)
	}

	books, err := lister.BooksForCharacter(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 {
		t.Fatalf("expected 2 books, got %d", len(books))
	}
	if books[0].Index != 1 || books[1].Index != 33 {
		t.Errorf("books out of order: %d, %d", books[0].Index, books[1].Index)
	}
}

func TestBooksForCharacter_BookName(t *testing.T) {
	dir := t.TempDir()
	// Write a book title for book 1 into mcr.ttl.
	ttlBuf := make([]byte, dat.HeaderSize+dat.MaxBooks/2*dat.BookNameSize)
	binary.LittleEndian.PutUint32(ttlBuf[0:4], dat.MagicVersion)
	copy(ttlBuf[dat.HeaderSize:dat.HeaderSize+dat.BookNameSize], []byte("WHM75NIN"))
	if err := os.WriteFile(filepath.Join(dir, "mcr.ttl"), ttlBuf, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), namedDat("Cure"), 0o644); err != nil {
		t.Fatal(err)
	}

	books, err := lister.BooksForCharacter(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].Name != "WHM75NIN" {
		t.Errorf("book name = %q, want WHM75NIN", books[0].Name)
	}
}

func TestBooksForCharacter_MissingDir(t *testing.T) {
	_, err := lister.BooksForCharacter("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}
