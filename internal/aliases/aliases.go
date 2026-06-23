package aliases

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const filename = "characters.yml"

// CurrentVersion is the highest characters.yml version this build understands.
const CurrentVersion = 1

// FutureVersionError is returned by Load when the file declares a version
// higher than CurrentVersion. The Document is still populated with whatever
// was parsed; callers may warn and continue for read-only use, but should
// refuse to write back to avoid stripping fields they don't understand.
type FutureVersionError struct {
	FileVersion int
}

func (e *FutureVersionError) Error() string {
	return fmt.Sprintf(
		"characters.yml declares version %d but this build only understands version %d; upgrade macromog before modifying aliases",
		e.FileVersion, CurrentVersion,
	)
}

// IsFutureVersion reports whether err is a FutureVersionError.
func IsFutureVersion(err error) bool {
	var fv *FutureVersionError
	return errors.As(err, &fv)
}

// Document is the top-level structure for USER/characters.yml.
type Document struct {
	Version int              `yaml:"version"`
	Chars   map[string]Entry `yaml:"chars,omitempty"`
}

// Entry holds the friendly alias for one character folder.
type Entry struct {
	Name string `yaml:"name"`
}

// Load reads characters.yml from userDir. Returns an empty Document with
// Version 1 if the file does not exist yet.
func Load(userDir string) (Document, error) {
	path := filepath.Join(userDir, filename)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Document{Version: 1}, nil
	}
	if err != nil {
		return Document{}, err
	}
	var doc Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Document{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if doc.Version > CurrentVersion {
		return doc, &FutureVersionError{FileVersion: doc.Version}
	}
	return doc, nil
}

// Save writes doc to characters.yml in userDir.
func Save(userDir string, doc Document) error {
	path := filepath.Join(userDir, filename)
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Resolve finds the hex folder ID for the given alias name (case-insensitive).
// Returns an error if name is empty, not found, or if multiple IDs share the name.
func Resolve(doc Document, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("character name must not be empty")
	}
	lower := strings.ToLower(name)
	var matches []string
	for id, entry := range doc.Chars {
		if strings.ToLower(entry.Name) == lower {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no character found with name %q", name)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple characters named %q: %v", name, matches)
	}
}

// LookupName returns the alias name for the given hex folder ID, or the ID
// itself if no alias is set.
func LookupName(doc Document, id string) string {
	if entry, ok := doc.Chars[id]; ok && entry.Name != "" {
		return entry.Name
	}
	return id
}
