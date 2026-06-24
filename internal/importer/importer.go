package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/squatched/macromog/internal/backup"
	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
	"github.com/squatched/macromog/internal/validate"
	"gopkg.in/yaml.v3"
)

// Options configures an import operation.
type Options struct {
	CharacterDir string
	YAMLPath     string
	// Scope overrides the scope embedded in the YAML. Zero = use YAML scope.
	Scope  scope.Scope
	Backup bool // create a timestamped backup before writing (default: true)
	DryRun bool // validate and plan without writing any files
}

// SetInfo describes one macro set that was (or would be) written.
type SetInfo struct {
	Book     int
	Set      int
	FileName string // e.g. "mcr320.dat"
}

// Result holds the outcome of an import operation.
type Result struct {
	BackupDir string    // path to the backup directory; empty when backup was skipped
	Sets      []SetInfo // sets written (or would-write for a dry run), sorted by book/set
}

type planEntry struct {
	book int
	set  int
	ms   dat.MacroSet
}

// Import reads opts.YAMLPath, validates it, optionally backs up the existing
// macro files, then writes the macro sets described in the YAML to opts.CharacterDir.
// Clearing of in-scope but absent entries is applied according to the effective scope.
func Import(opts Options) (Result, error) {
	data, err := os.ReadFile(opts.YAMLPath)
	if err != nil {
		return Result{}, fmt.Errorf("read %s: %w", opts.YAMLPath, err)
	}

	if errs := validate.Validate(data); len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = "  " + e.Error()
		}
		return Result{}, fmt.Errorf("validation failed:\n%s", strings.Join(msgs, "\n"))
	}

	var doc export.Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Result{}, fmt.Errorf("parse YAML: %w", err)
	}

	// Determine effective scope: --scope flag takes precedence over YAML field.
	sc := opts.Scope
	if sc.IsZero() {
		sc = doc.Scope
	}
	// sc.IsZero() after this = legacy YAML with no scope field = write-only (no clearing).

	// Build the sorted write plan.
	var plan []planEntry

	if sc.Level == scope.LevelMacro {
		// Macro scope: read existing sets, merge only scoped slots, write back.
		plan, err = buildMacroScopePlan(opts.CharacterDir, doc, sc)
		if err != nil {
			return Result{}, err
		}
	} else {
		// Full / book / set scope (or legacy): write full MacroSets from YAML.
		// When an explicit scope is active, only write books/sets within it.
		for bookIdx, book := range doc.Books {
			if !sc.IsZero() && !sc.ContainsBook(bookIdx) {
				continue
			}
			for setIdx, s := range book.Sets {
				if !sc.IsZero() && !sc.ContainsSet(bookIdx, setIdx) {
					continue
				}
				plan = append(plan, planEntry{book: bookIdx, set: setIdx, ms: buildMacroSet(s)})
			}
		}
	}

	sort.Slice(plan, func(i, j int) bool {
		if plan[i].book != plan[j].book {
			return plan[i].book < plan[j].book
		}
		return plan[i].set < plan[j].set
	})

	sets := make([]SetInfo, len(plan))
	for i, p := range plan {
		sets[i] = SetInfo{Book: p.book, Set: p.set, FileName: dat.MacroFileName(p.book, p.set)}
	}

	if opts.DryRun {
		return Result{Sets: sets}, nil
	}

	// Backup before any writes.
	var backupDir string
	if opts.Backup {
		backupDir, err = backup.Backup(opts.CharacterDir, filepath.Join(opts.CharacterDir, "backups"))
		if err != nil {
			return Result{BackupDir: backupDir, Sets: sets}, fmt.Errorf("backup: %w", err)
		}
	}

	// Write macro .dat files (content from YAML).
	for _, p := range plan {
		path := filepath.Join(opts.CharacterDir, dat.MacroFileName(p.book, p.set))
		if err := dat.WriteMacroSetFile(path, p.ms); err != nil {
			return Result{BackupDir: backupDir, Sets: sets}, fmt.Errorf("write %s: %w", dat.MacroFileName(p.book, p.set), err)
		}
	}

	// Clear in-scope entries absent from the YAML.
	if err := applyClearing(opts.CharacterDir, doc, sc); err != nil {
		return Result{BackupDir: backupDir, Sets: sets}, fmt.Errorf("clear: %w", err)
	}

	// Update book title files (scope-aware).
	if err := updateBookTitles(opts.CharacterDir, doc, sc); err != nil {
		return Result{BackupDir: backupDir, Sets: sets}, fmt.Errorf("update book titles: %w", err)
	}

	return Result{BackupDir: backupDir, Sets: sets}, nil
}

// applyClearing deletes or zeros DAT files for in-scope entries absent from the YAML.
// Legacy (zero scope) and macro scope skip this phase entirely.
func applyClearing(dir string, doc export.Document, sc scope.Scope) error {
	if sc.IsZero() || sc.Level == scope.LevelMacro {
		return nil
	}

	if sc.Level == scope.LevelSet {
		return clearSetScope(dir, doc, sc)
	}

	// Full or book scope: iterate all books in scope.
	for _, bookIdx := range sc.BooksInScope(dat.MaxBooks) {
		yamlBook, inYAML := doc.Books[bookIdx]
		if !inYAML {
			// Entire book absent → delete all its .dat files.
			if err := deleteBookDatFiles(dir, bookIdx); err != nil {
				return err
			}
			continue
		}
		// Book present in YAML: delete absent sets' .dat files.
		for setIdx := 1; setIdx <= dat.SetsPerBook; setIdx++ {
			if _, hasSet := yamlBook.Sets[setIdx]; !hasSet {
				if err := deleteIfExists(filepath.Join(dir, dat.MacroFileName(bookIdx, setIdx))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// clearSetScope deletes DAT files for (book, set) pairs that are in scope but absent from YAML.
func clearSetScope(dir string, doc export.Document, sc scope.Scope) error {
	for _, sel := range sc.Selections {
		yamlBook, bookInYAML := doc.Books[sel.Book]
		if !bookInYAML {
			if err := deleteIfExists(filepath.Join(dir, dat.MacroFileName(sel.Book, sel.Set))); err != nil {
				return err
			}
			continue
		}
		if _, setInYAML := yamlBook.Sets[sel.Set]; !setInYAML {
			if err := deleteIfExists(filepath.Join(dir, dat.MacroFileName(sel.Book, sel.Set))); err != nil {
				return err
			}
		}
	}
	return nil
}

// deleteBookDatFiles removes all mcr*.dat files for the given book index.
func deleteBookDatFiles(dir string, book int) error {
	for setIdx := 1; setIdx <= dat.SetsPerBook; setIdx++ {
		if err := deleteIfExists(filepath.Join(dir, dat.MacroFileName(book, setIdx))); err != nil {
			return err
		}
	}
	return nil
}

func deleteIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete %s: %w", filepath.Base(path), err)
	}
	return nil
}

// buildMacroScopePlan reads existing MacroSets and merges only the scoped slots from the YAML.
func buildMacroScopePlan(dir string, doc export.Document, sc scope.Scope) ([]planEntry, error) {
	type bookSet struct{ book, set int }
	seen := make(map[bookSet]bool)
	var order []bookSet
	for _, sel := range sc.Selections {
		bs := bookSet{sel.Book, sel.Set}
		if !seen[bs] {
			seen[bs] = true
			order = append(order, bs)
		}
	}

	var entries []planEntry
	for _, bs := range order {
		path := filepath.Join(dir, dat.MacroFileName(bs.book, bs.set))
		existing, err := dat.ReadMacroSetFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read %s: %w", dat.MacroFileName(bs.book, bs.set), err)
		}
		if yamlBook, ok := doc.Books[bs.book]; ok {
			if yamlSet, ok := yamlBook.Sets[bs.set]; ok {
				mergeMacroScopeSlots(&existing, yamlSet, sc, bs.book, bs.set)
			}
		}
		entries = append(entries, planEntry{book: bs.book, set: bs.set, ms: existing})
	}
	return entries, nil
}

// mergeMacroScopeSlots writes only the slots that appear in both sc and the YAML set.
func mergeMacroScopeSlots(ms *dat.MacroSet, s export.Set, sc scope.Scope, book, set int) {
	for yamlKey, m := range s.Ctrl {
		if sc.ContainsMacro(book, set, scope.TypeCtrl, yamlKey) {
			slot := dat.SlotFromYAMLKey(yamlKey)
			if slot >= 0 && slot < dat.SetsPerBook {
				ms.Ctrl[slot] = buildMacro(m)
			}
		}
	}
	for yamlKey, m := range s.Alt {
		if sc.ContainsMacro(book, set, scope.TypeAlt, yamlKey) {
			slot := dat.SlotFromYAMLKey(yamlKey)
			if slot >= 0 && slot < dat.SetsPerBook {
				ms.Alt[slot] = buildMacro(m)
			}
		}
	}
}

// updateBookTitles reads existing titles, applies YAML names within scope, and writes back.
// Books within scope but absent from the YAML get their title cleared.
func updateBookTitles(dir string, doc export.Document, sc scope.Scope) error {
	titles, err := dat.ReadBookTitles(dir)
	if err != nil {
		return err
	}

	if sc.IsZero() {
		// Legacy: only write titles that appear in the YAML, no clearing.
		for bookIdx, book := range doc.Books {
			if bookIdx >= 1 && bookIdx <= dat.MaxBooks {
				titles[bookIdx-1] = book.Name
			}
		}
		return dat.WriteBookTitles(dir, titles)
	}

	// Within scope: apply YAML name or clear title for absent books.
	for _, bookIdx := range sc.BooksInScope(dat.MaxBooks) {
		if book, ok := doc.Books[bookIdx]; ok {
			titles[bookIdx-1] = book.Name
		} else if sc.Level != scope.LevelMacro {
			// Book absent from YAML within scope: clear title
			// (for set/macro scope, leave title untouched for non-empty books)
			if sc.Level == scope.LevelFull || sc.Level == scope.LevelBook {
				titles[bookIdx-1] = ""
			}
		}
	}

	return dat.WriteBookTitles(dir, titles)
}

func buildMacroSet(s export.Set) dat.MacroSet {
	ms := dat.MacroSet{HeaderUnknown: s.HeaderUnknown}
	for yamlKey, m := range s.Ctrl {
		slot := dat.SlotFromYAMLKey(yamlKey)
		if slot >= 0 && slot < dat.SetsPerBook {
			ms.Ctrl[slot] = buildMacro(m)
		}
	}
	for yamlKey, m := range s.Alt {
		slot := dat.SlotFromYAMLKey(yamlKey)
		if slot >= 0 && slot < dat.SetsPerBook {
			ms.Alt[slot] = buildMacro(m)
		}
	}
	return ms
}

func buildMacro(m export.Macro) dat.Macro {
	dm := dat.Macro{Name: m.Name}
	for i, line := range m.Contents {
		if i < dat.LineCount {
			dm.Contents[i] = line
		}
	}
	return dm
}
