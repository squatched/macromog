package dat

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var macroFileRe = regexp.MustCompile(`^mcr(\d*)\.dat$`)

// FileIndex returns the combined book/set index used in mcr*.dat filenames.
// Books and sets are zero-based: book 1 set 1 => index 0 (mcr.dat).
func FileIndex(book, set int) int {
	return (book-1)*SetsPerBook + (set - 1)
}

// ParseFileIndex splits a combined index into 1-based book and set numbers.
func ParseFileIndex(index int) (book, set int) {
	book = index/SetsPerBook + 1
	set = index%SetsPerBook + 1
	return book, set
}

// MacroFileName returns the on-disk filename for a book/set pair.
// Leading zeros are dropped: book 1 set 1 => "mcr.dat", book 1 set 2 => "mcr1.dat".
func MacroFileName(book, set int) string {
	index := FileIndex(book, set)
	if index == 0 {
		return "mcr.dat"
	}
	return fmt.Sprintf("mcr%d.dat", index)
}

// ParseMacroFileName extracts the combined index from a filename like mcr320.dat.
// Returns ok=false for non-macro files (e.g. nmcr.dat, mcr.ttl).
func ParseMacroFileName(name string) (index int, ok bool) {
	base := strings.ToLower(filepath.Base(name))
	m := macroFileRe.FindStringSubmatch(base)
	if m == nil {
		return 0, false
	}
	if m[1] == "" {
		return 0, true
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 0 || n >= MaxFileIndex {
		return 0, false
	}
	return n, true
}

// YAMLKey converts a zero-based ctrl/alt slot (0-9) to the YAML key index
// used by macromog: 1-9, then 0 for the bottom-row key.
func YAMLKey(slot int) int {
	if slot == 9 {
		return 0
	}
	return slot + 1
}