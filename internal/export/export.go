package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/squatched/macromog/internal/dat"
	"gopkg.in/yaml.v3"
)

// Document is the macromog YAML export schema (sparse).
type Document struct {
	Version    int          `yaml:"version"`
	Character  string       `yaml:"character,omitempty"`
	ExportedAt string       `yaml:"exported_at,omitempty"`
	Books      map[int]Book `yaml:"books,omitempty"`
}

type Book struct {
	Name string         `yaml:"name,omitempty"`
	Sets map[int]Set    `yaml:"sets,omitempty"`
}

type Set struct {
	Ctrl map[int]Macro `yaml:"ctrl,omitempty"`
	Alt  map[int]Macro `yaml:"alt,omitempty"`
}

type Macro struct {
	Name     string   `yaml:"name,omitempty"`
	Contents []string `yaml:"contents,omitempty"`
}

// Options configures an export run.
type Options struct {
	CharacterDir string
	Character    string
}

// FromCharacterDir reads macro .dat files from a FFXI USER directory and
// returns a sparse YAML document containing all non-empty macros.
func FromCharacterDir(opts Options) (Document, error) {
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
		Books:      make(map[int]Book),
	}

	for _, path := range files {
		index, ok := dat.ParseMacroFileName(filepath.Base(path))
		if !ok {
			continue
		}
		book, set := dat.ParseFileIndex(index)

		setData, err := dat.ReadMacroSetFile(path)
		if err != nil {
			return Document{}, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}

		exported := exportMacroSet(setData)
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

	if len(doc.Books) == 0 {
		return Document{}, fmt.Errorf("no macro data found in %s", dir)
	}

	return doc, nil
}

func exportMacroSet(set dat.MacroSet) Set {
	out := Set{}
	for i := 0; i < 10; i++ {
		if m := exportMacro(set.Ctrl[i]); m != nil {
			if out.Ctrl == nil {
				out.Ctrl = make(map[int]Macro)
			}
			out.Ctrl[dat.YAMLKey(i)] = *m
		}
		if m := exportMacro(set.Alt[i]); m != nil {
			if out.Alt == nil {
				out.Alt = make(map[int]Macro)
			}
			out.Alt[dat.YAMLKey(i)] = *m
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
	node, err := buildYAMLNode(doc)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(node)
}

func buildYAMLNode(doc Document) (*yaml.Node, error) {
	root := &yaml.Node{Kind: yaml.MappingNode}
	addKV(root, "version", intNode(doc.Version))
	if doc.Character != "" {
		addKV(root, "character", scalarNode(doc.Character))
	}
	if doc.ExportedAt != "" {
		addKV(root, "exported_at", scalarNode(doc.ExportedAt))
	}

	bookKeys := sortedKeys(doc.Books)
	if len(bookKeys) > 0 {
		booksNode := &yaml.Node{Kind: yaml.MappingNode}
		for _, bk := range bookKeys {
			book := doc.Books[bk]
			bookNode := &yaml.Node{Kind: yaml.MappingNode}
			if book.Name != "" {
				addKV(bookNode, "name", scalarNode(book.Name))
			}
			setKeys := sortedKeys(book.Sets)
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

	return root, nil
}

func setNode(s Set) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
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
	for _, key := range sortedKeys(row) {
		addKV(n, key, macroNode(row[key]))
	}
	return n
}

func macroNode(m Macro) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	if m.Name != "" {
		addKV(n, "name", scalarNode(m.Name))
	}
	if len(m.Contents) > 0 {
		lines := &yaml.Node{Kind: yaml.SequenceNode}
		for _, line := range m.Contents {
			lines.Content = append(lines.Content, scalarNode(line))
		}
		addKV(n, "contents", lines)
	}
	return n
}

func addKV(parent *yaml.Node, key interface{}, value *yaml.Node) {
	keyNode := keyNode(key)
	parent.Content = append(parent.Content, keyNode, value)
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
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func intNode(value int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", value)}
}

func sortedKeys[V any](m map[int]V) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// WriteFile exports macros from characterDir and writes YAML to outputPath.
func WriteFile(characterDir, outputPath, character string) error {
	doc, err := FromCharacterDir(Options{CharacterDir: characterDir, Character: character})
	if err != nil {
		return err
	}
	data, err := MarshalYAML(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o644)
}
