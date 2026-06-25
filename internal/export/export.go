package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/scope"
	"gopkg.in/yaml.v3"
)

// Document is the macromog YAML export schema (sparse).
type Document struct {
	Version    int          `yaml:"version"`
	Character  string       `yaml:"character,omitempty"`
	ExportedAt string       `yaml:"exported_at,omitempty"`
	Scope      scope.Scope  `yaml:"scope"`
	Books      map[int]Book `yaml:"books,omitempty"`
}

type Book struct {
	Name string      `yaml:"name,omitempty"`
	Sets map[int]Set `yaml:"sets,omitempty"`
}

type Set struct {
	HeaderUnknown uint32        `yaml:"header_unknown,omitempty"`
	Ctrl          map[int]Macro `yaml:"ctrl,omitempty"`
	Alt           map[int]Macro `yaml:"alt,omitempty"`
}

type Macro struct {
	Name     string   `yaml:"name,omitempty"`
	Contents []string `yaml:"contents,omitempty"`
}

// Options configures an export run.
type Options struct {
	CharacterDir string
	Character    string
	Scope        scope.Scope // zero value = full scope
	Dense        bool        // include all in-scope macros even if empty
}

// FromCharacterDir reads macro .dat files from a FFXI USER directory and
// returns a YAML document containing macros within the scope. When
// opts.Dense is false (default), only non-empty macros are included. When
// opts.Dense is true, all in-scope slots are present even when empty.
func FromCharacterDir(opts Options) (Document, error) {
	sc := opts.Scope
	if sc.IsZero() {
		sc = scope.Full()
	}

	if opts.Dense {
		return fromCharacterDirDense(opts, sc)
	}

	dir := opts.CharacterDir
	titles, err := dat.ReadBookTitles(dir)
	if err != nil {
		return Document{}, err
	}

	files, err := dat.DiscoverMacroFiles(dir)
	if err != nil {
		return Document{}, err
	}

	doc := Document{
		Version:    1,
		Character:  opts.Character,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Scope:      sc,
		Books:      make(map[int]Book),
	}

	for _, path := range files {
		index, ok := dat.ParseMacroFileName(filepath.Base(path))
		if !ok {
			continue
		}
		book, set := dat.ParseFileIndex(index)
		if book < 1 || book > dat.MaxBooks {
			continue
		}

		// Skip book/set pairs outside the scope.
		if !sc.ContainsSet(book, set) {
			continue
		}

		setData, err := dat.ReadMacroSetFile(path)
		if err != nil {
			return Document{}, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}

		exported := exportMacroSet(setData, sc, book, set)
		if exported.Ctrl == nil && exported.Alt == nil {
			continue
		}

		b, ok := doc.Books[book]
		if !ok {
			b = Book{Sets: make(map[int]Set)}
			if name := titles[book-1]; name != "" {
				b.Name = name
			}
		}
		b.Sets[set] = exported
		doc.Books[book] = b
	}

	return doc, nil
}

func exportMacroSet(set dat.MacroSet, sc scope.Scope, book, setIdx int) Set {
	out := Set{HeaderUnknown: set.HeaderUnknown}
	for i := 0; i < 10; i++ {
		yamlKey := dat.YAMLKey(i)
		if sc.ContainsMacro(book, setIdx, scope.TypeCtrl, yamlKey) {
			if m := exportMacro(set.Ctrl[i]); m != nil {
				if out.Ctrl == nil {
					out.Ctrl = make(map[int]Macro)
				}
				out.Ctrl[yamlKey] = *m
			}
		}
		if sc.ContainsMacro(book, setIdx, scope.TypeAlt, yamlKey) {
			if m := exportMacro(set.Alt[i]); m != nil {
				if out.Alt == nil {
					out.Alt = make(map[int]Macro)
				}
				out.Alt[yamlKey] = *m
			}
		}
	}
	return out
}

func exportMacro(m dat.Macro) *Macro {
	if m.Empty() {
		return nil
	}
	out := Macro{}
	if m.Name != "" {
		out.Name = m.Name
	}
	lines := compactLines(m.Contents)
	if len(lines) > 0 {
		out.Contents = lines
	}
	return &out
}

// fromCharacterDirDense builds a Document that includes every in-scope book,
// set, and macro slot regardless of whether the source files or content exist.
// Missing .dat files are treated as empty sets.
func fromCharacterDirDense(opts Options, sc scope.Scope) (Document, error) {
	dir := opts.CharacterDir
	titles, err := dat.ReadBookTitles(dir)
	if err != nil {
		return Document{}, err
	}

	doc := Document{
		Version:    1,
		Character:  opts.Character,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Scope:      sc,
		Books:      make(map[int]Book),
	}

	for _, bookIdx := range sc.BooksInScope(dat.MaxBooks) {
		b := Book{Sets: make(map[int]Set)}
		if name := titles[bookIdx-1]; name != "" {
			b.Name = name
		}
		for setIdx := 1; setIdx <= dat.SetsPerBook; setIdx++ {
			if !sc.ContainsSet(bookIdx, setIdx) {
				continue
			}
			path := filepath.Join(dir, dat.MacroFileName(bookIdx, setIdx))
			setData, readErr := dat.ReadMacroSetFile(path)
			if readErr != nil && !os.IsNotExist(readErr) {
				return Document{}, fmt.Errorf("%s: %w", dat.MacroFileName(bookIdx, setIdx), readErr)
			}
			b.Sets[setIdx] = exportMacroSetDense(setData, sc, bookIdx, setIdx)
		}
		doc.Books[bookIdx] = b
	}

	return doc, nil
}

func exportMacroSetDense(set dat.MacroSet, sc scope.Scope, book, setIdx int) Set {
	out := Set{HeaderUnknown: set.HeaderUnknown}
	for i := 0; i < dat.SetsPerBook; i++ {
		yamlKey := dat.YAMLKey(i)
		if sc.ContainsMacro(book, setIdx, scope.TypeCtrl, yamlKey) {
			if out.Ctrl == nil {
				out.Ctrl = make(map[int]Macro)
			}
			out.Ctrl[yamlKey] = exportMacroDense(set.Ctrl[i])
		}
		if sc.ContainsMacro(book, setIdx, scope.TypeAlt, yamlKey) {
			if out.Alt == nil {
				out.Alt = make(map[int]Macro)
			}
			out.Alt[yamlKey] = exportMacroDense(set.Alt[i])
		}
	}
	return out
}

// exportMacroDense always returns a Macro with all dat.LineCount content lines
// populated. This gives the post-processor (replacePlaceholders) all 6 slots
// to work with; empty strings become numbered comment placeholders.
func exportMacroDense(m dat.Macro) Macro {
	out := Macro{}
	if m.Name != "" {
		out.Name = m.Name
	}
	full := make([]string, dat.LineCount)
	copy(full, m.Contents[:])
	out.Contents = full
	return out
}

// compactLines preserves interior blank lines but trims trailing empties.
func compactLines(lines [dat.LineCount]string) []string {
	last := -1
	for i := dat.LineCount - 1; i >= 0; i-- {
		if lines[i] != "" {
			last = i
			break
		}
	}
	if last < 0 {
		return nil
	}
	out := make([]string, last+1)
	for i := 0; i <= last; i++ {
		out[i] = lines[i]
	}
	return out
}

// MarshalYAML encodes doc with stable key ordering for maps.
func MarshalYAML(doc Document) ([]byte, error) {
	return yaml.Marshal(buildYAMLNode(doc))
}

func buildYAMLNode(doc Document) *yaml.Node {
	root := &yaml.Node{Kind: yaml.MappingNode}
	addKV(root, "version", intNode(doc.Version))
	if doc.Character != "" {
		addKV(root, "character", scalarNode(doc.Character))
	}
	if doc.ExportedAt != "" {
		addKV(root, "exported_at", scalarNode(doc.ExportedAt))
	}

	// scope is always written.
	if !doc.Scope.IsZero() {
		addKV(root, "scope", scopeNode(doc.Scope))
	}

	bookKeys := sortedIntKeys(doc.Books)
	if len(bookKeys) > 0 {
		booksNode := &yaml.Node{Kind: yaml.MappingNode}
		for _, bk := range bookKeys {
			book := doc.Books[bk]
			bookNode := &yaml.Node{Kind: yaml.MappingNode}
			if book.Name != "" {
				addKV(bookNode, "name", scalarNode(book.Name))
			}
			setKeys := sortedIntKeys(book.Sets)
			if len(setKeys) > 0 {
				setsNode := &yaml.Node{Kind: yaml.MappingNode}
				for _, sk := range setKeys {
					addKV(setsNode, sk, setNode(book.Sets[sk]))
				}
				addKV(bookNode, "sets", setsNode)
			}
			addKV(booksNode, bk, bookNode)
		}
		addKV(root, "books", booksNode)
	}

	return root
}

func scopeNode(sc scope.Scope) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	addKV(n, "level", scalarNode(string(sc.Level)))
	if len(sc.Selections) > 0 {
		seqNode := &yaml.Node{Kind: yaml.SequenceNode}
		for _, sel := range sc.Selections {
			seqNode.Content = append(seqNode.Content, selectionNode(sel))
		}
		addKV(n, "selections", seqNode)
	}
	return n
}

func selectionNode(sel scope.Selection) *yaml.Node {
	// Use flow style (inline) for compact output: {book: 1, set: 3, type: ctrl, key: 1}
	n := &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle}
	addKV(n, "book", intNode(sel.Book))
	if sel.Set != 0 {
		addKV(n, "set", intNode(sel.Set))
	}
	if sel.Type != "" {
		addKV(n, "type", scalarNode(string(sel.Type)))
	}
	if sel.Type != "" {
		addKV(n, "key", intNode(sel.Key))
	}
	return n
}

func setNode(s Set) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	if s.HeaderUnknown != 0 {
		addKV(n, "header_unknown", uint32Node(s.HeaderUnknown))
	}
	if len(s.Ctrl) > 0 {
		addKV(n, "ctrl", macroRowNode(s.Ctrl))
	}
	if len(s.Alt) > 0 {
		addKV(n, "alt", macroRowNode(s.Alt))
	}
	return n
}

func macroRowNode(row map[int]Macro) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	for _, key := range sortedIntKeys(row) {
		addKV(n, key, macroNode(row[key]))
	}
	return n
}

func macroNode(m Macro) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	if m.Name != "" {
		addKV(n, "name", scalarNode(m.Name))
	}
	if m.Contents != nil {
		lines := &yaml.Node{Kind: yaml.SequenceNode}
		for _, line := range m.Contents {
			lines.Content = append(lines.Content, scalarNode(line))
		}
		addKV(n, "contents", lines)
	}
	return n
}

func addKV(parent *yaml.Node, key interface{}, value *yaml.Node) {
	kn := keyNode(key)
	parent.Content = append(parent.Content, kn, value)
}

func keyNode(key interface{}) *yaml.Node {
	switch k := key.(type) {
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", k)}
	default:
		return scalarNode(fmt.Sprint(k))
	}
}

func scalarNode(value string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode, Value: value}
	if value == "" {
		// A bare scalar node with no value serializes as YAML null, which
		// yaml.v3 drops when unmarshaling into []string. Use double-quoted
		// style so empty strings round-trip as "" instead of vanishing.
		n.Style = yaml.DoubleQuotedStyle
	}
	return n
}

func intNode(value int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", value)}
}

func uint32Node(value uint32) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", value)}
}

func sortedIntKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// MarshalYAMLWithPlaceholders is like MarshalYAML but replaces double-quoted
// empty-string items within "contents:" blocks with numbered comment lines
// (# Macro Line N). This prevents users from being misled into wrapping their
// macro content in outer double quotes when editing the YAML.
//
// "contents: []" (inline empty sequence) is left unchanged — it signals a
// named macro with no content lines, not an unfilled placeholder slot.
func MarshalYAMLWithPlaceholders(doc Document) ([]byte, error) {
	data, err := MarshalYAML(doc)
	if err != nil {
		return nil, err
	}
	return replacePlaceholders(data), nil
}

// replacePlaceholders processes YAML line-by-line. Within each block-style
// "contents:" sequence it replaces every `- ""` item with a numbered comment.
// Each item is numbered by its 1-based position in the sequence regardless of
// whether surrounding items are empty or non-empty.
func replacePlaceholders(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))

	itemNum := 0
	itemIndent := -1 // indent of sequence items; -1 = not in a contents block

	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			// Preserve the trailing newline artifact from Split.
			out = append(out, line)
			continue
		}

		stripped := strings.TrimLeft(line, " ")
		indent := len(line) - len(stripped)

		if itemIndent >= 0 {
			if indent == itemIndent && strings.HasPrefix(stripped, "- ") {
				itemNum++
				if stripped == `- ""` {
					out = append(out, fmt.Sprintf("%s- # Macro Line %d",
						strings.Repeat(" ", itemIndent), itemNum))
				} else {
					out = append(out, line)
				}
				continue
			}
			// Indentation changed or non-item line: exit the contents block.
			itemIndent = -1
			itemNum = 0
		}

		// Detect the start of a block-style contents: sequence.
		// "contents:" (bare key) starts a block; "contents: []" does not.
		if stripped == "contents:" {
			itemIndent = indent + 4
			itemNum = 0
		}

		out = append(out, line)
	}

	return []byte(strings.Join(out, "\n"))
}

// WriteFile exports macros from opts.CharacterDir and writes YAML to outputPath.
// WriteTo writes an export document to w.
func WriteTo(w io.Writer, opts Options) error {
	doc, err := FromCharacterDir(opts)
	if err != nil {
		return err
	}
	data, err := MarshalYAMLWithPlaceholders(doc)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func WriteFile(opts Options, outputPath string) error {
	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if err := WriteTo(f, opts); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
