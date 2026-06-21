package validate

import (
	"fmt"
	"regexp"
	"sort"
	"time"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Error is a single schema violation with its location and description.
type Error struct {
	Path    string
	Message string
}

func (e Error) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return e.Path + ": " + e.Message
}

// document is the top-level YAML structure.
// Version uses a pointer so a nil value distinguishes absent from zero.
type document struct {
	Version    *int         `yaml:"version"`
	Character  string       `yaml:"character"`
	ExportedAt string       `yaml:"exported_at"`
	Books      map[int]book `yaml:"books"`
}

type book struct {
	Name string      `yaml:"name"`
	Sets map[int]set `yaml:"sets"`
}

type set struct {
	Ctrl map[int]macro `yaml:"ctrl"`
	Alt  map[int]macro `yaml:"alt"`
}

type macro struct {
	Name     string   `yaml:"name"`
	Contents []string `yaml:"contents"`
}

var (
	bookNameRe      = regexp.MustCompile(`^[A-Za-z0-9]*$`)
	resourceMarkerRe = regexp.MustCompile(`≺\[[0-9A-Fa-f]{8}]≻`)
)

func containsResourceMarker(s string) bool {
	return resourceMarkerRe.MatchString(s)
}

func hasNonPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}

// Validate parses data as YAML and checks it against the macromog schema.
// Returns an empty slice when the file is valid.
func Validate(data []byte) []Error {
	var errs []Error

	var doc document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return []Error{{Message: "invalid YAML: " + err.Error()}}
	}

	if doc.Version == nil {
		errs = append(errs, Error{Path: "version", Message: "required field missing"})
	} else if *doc.Version != 1 {
		errs = append(errs, Error{Path: "version", Message: fmt.Sprintf("must be 1, got %d", *doc.Version)})
	}

	if doc.Books == nil {
		errs = append(errs, Error{Path: "books", Message: "required field missing"})
	}

	if doc.ExportedAt != "" {
		if _, err := time.Parse(time.RFC3339, doc.ExportedAt); err != nil {
			errs = append(errs, Error{Path: "exported_at", Message: `must be a date-time like "2026-06-20T03:30:00Z"`})
		}
	}

	for bookIdx, b := range doc.Books {
		errs = append(errs, validateBook(bookIdx, b)...)
	}

	sort.Slice(errs, func(i, j int) bool { return errs[i].Path < errs[j].Path })

	return errs
}

func validateBook(idx int, b book) []Error {
	var errs []Error
	path := fmt.Sprintf("books.%d", idx)

	if idx < 1 || idx > 40 {
		errs = append(errs, Error{Path: path, Message: fmt.Sprintf("index must be 1–40, got %d", idx)})
	}

	if len(b.Name) > 15 {
		errs = append(errs, Error{Path: path + ".name", Message: fmt.Sprintf("max 15 characters, %q is %d", b.Name, len(b.Name))})
	}
	if b.Name != "" && !bookNameRe.MatchString(b.Name) {
		errs = append(errs, Error{Path: path + ".name", Message: fmt.Sprintf("only alphanumeric characters allowed, got %q", b.Name)})
	}

	for setIdx, s := range b.Sets {
		errs = append(errs, validateSet(path, setIdx, s)...)
	}

	return errs
}

func validateSet(bookPath string, idx int, s set) []Error {
	var errs []Error
	path := fmt.Sprintf("%s.sets.%d", bookPath, idx)

	if idx < 1 || idx > 10 {
		errs = append(errs, Error{Path: path, Message: fmt.Sprintf("index must be 1–10, got %d", idx)})
	}

	errs = append(errs, validateMacroRow(path+".ctrl", s.Ctrl)...)
	errs = append(errs, validateMacroRow(path+".alt", s.Alt)...)

	return errs
}

func validateMacroRow(path string, row map[int]macro) []Error {
	var errs []Error

	for key, m := range row {
		macroPath := fmt.Sprintf("%s.%d", path, key)

		if key < 0 || key > 9 {
			errs = append(errs, Error{Path: macroPath, Message: fmt.Sprintf("key must be 0–9, got %d", key)})
		}

		errs = append(errs, validateMacro(macroPath, m)...)
	}

	return errs
}

func validateMacro(path string, m macro) []Error {
	var errs []Error

	if nameLen := utf8.RuneCountInString(m.Name); nameLen > 8 {
		errs = append(errs, Error{Path: path + ".name", Message: fmt.Sprintf("max 8 characters, %q is %d", m.Name, nameLen)})
	}
	if m.Name != "" && hasNonPrintable(m.Name) {
		errs = append(errs, Error{Path: path + ".name", Message: fmt.Sprintf("must contain only printable characters, got %q", m.Name)})
	}

	if len(m.Contents) > 6 {
		errs = append(errs, Error{Path: path + ".contents", Message: fmt.Sprintf("max 6 lines, got %d", len(m.Contents))})
	}

	for i, line := range m.Contents {
		linePath := fmt.Sprintf("%s.contents[%d]", path, i)
		// Auto-translate resource markers (≺[XXXXXXXX]≻) expand in-game; YAML
		// length cannot reflect the client-side character budget.
		if !containsResourceMarker(line) {
			if lineLen := utf8.RuneCountInString(line); lineLen > 60 {
				errs = append(errs, Error{Path: linePath, Message: fmt.Sprintf("max 60 characters, %q is %d", line, lineLen)})
			}
		}
		if hasNonPrintable(line) {
			errs = append(errs, Error{Path: linePath, Message: fmt.Sprintf("must contain only printable characters, got %q", line)})
		}
	}

	return errs
}
