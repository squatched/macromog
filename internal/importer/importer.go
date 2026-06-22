package importer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/validate"
	"gopkg.in/yaml.v3"
)

// Options configures an import operation.
type Options struct {
	CharacterDir string
	YAMLPath     string
	Backup       bool // create a timestamped backup before writing (default: true)
	DryRun       bool // validate and plan without writing any files
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

// Import reads opts.YAMLPath, validates it, optionally backs up the existing
// macro files, then writes the macro sets described in the YAML to opts.CharacterDir.
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

	// Build the sorted write plan.
	type planEntry struct {
		book int
		set  int
		ms   dat.MacroSet
	}
	var plan []planEntry
	for bookIdx, book := range doc.Books {
		for setIdx, s := range book.Sets {
			plan = append(plan, planEntry{bookIdx, setIdx, buildMacroSet(s)})
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
		backupDir, err = backupCharDir(opts.CharacterDir)
		if err != nil {
			return Result{}, fmt.Errorf("backup: %w", err)
		}
	}

	// Write macro .dat files.
	for _, p := range plan {
		path := filepath.Join(opts.CharacterDir, dat.MacroFileName(p.book, p.set))
		if err := dat.WriteMacroSetFile(path, p.ms); err != nil {
			return Result{BackupDir: backupDir, Sets: sets}, fmt.Errorf("write %s: %w", dat.MacroFileName(p.book, p.set), err)
		}
	}

	// Update book title files.
	if err := updateBookTitles(opts.CharacterDir, doc); err != nil {
		return Result{BackupDir: backupDir, Sets: sets}, fmt.Errorf("update book titles: %w", err)
	}

	return Result{BackupDir: backupDir, Sets: sets}, nil
}

// backupCharDir copies all *.dat and *.ttl files in dir to a timestamped
// subdirectory and returns its path.
func backupCharDir(dir string) (string, error) {
	stamp := time.Now().UTC().Format("20060102_150405")
	backupDir := filepath.Join(dir, "backups", stamp)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if !strings.HasSuffix(lower, ".dat") && !strings.HasSuffix(lower, ".ttl") {
			continue
		}
		if err := copyFile(filepath.Join(dir, e.Name()), filepath.Join(backupDir, e.Name())); err != nil {
			return "", err
		}
	}
	return backupDir, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// updateBookTitles reads existing book titles, applies any names from doc,
// then writes both .ttl files.
func updateBookTitles(dir string, doc export.Document) error {
	titles, err := dat.ReadBookTitles(dir)
	if err != nil {
		return err
	}
	for bookIdx, book := range doc.Books {
		if bookIdx >= 1 && bookIdx <= dat.MaxBooks {
			titles[bookIdx-1] = book.Name
		}
	}
	return dat.WriteBookTitles(dir, titles)
}

func buildMacroSet(s export.Set) dat.MacroSet {
	ms := dat.MacroSet{HeaderUnknown: s.HeaderUnknown}
	for yamlKey, m := range s.Ctrl {
		slot := dat.SlotFromYAMLKey(yamlKey)
		if slot >= 0 && slot < 10 {
			ms.Ctrl[slot] = buildMacro(m)
		}
	}
	for yamlKey, m := range s.Alt {
		slot := dat.SlotFromYAMLKey(yamlKey)
		if slot >= 0 && slot < 10 {
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
