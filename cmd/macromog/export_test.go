package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestRunExport_MissingChar(t *testing.T) {
	if got := runExport(nil, newTextPrinter()); got != 1 {
		t.Errorf("runExport(nil) = %d, want 1", got)
	}
}

func TestRunExport_Help(t *testing.T) {
	if got := runExport([]string{"--help"}, newTextPrinter()); got != 0 {
		t.Errorf("runExport(--help) = %d, want 0", got)
	}
}

func TestRunExport_BadCharDir(t *testing.T) {
	if got := runExport([]string{"--char-dir", "/nonexistent/char"}, newTextPrinter()); got != 1 {
		t.Errorf("runExport(bad char) = %d, want 1", got)
	}
}

func TestRunExport_Book33(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "book33.yml")
	args := []string{"--char-dir", testdata.CharDir(), out}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Errorf("runExport = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "B33S1") {
		t.Errorf("missing B33S1 in output: %s", data)
	}
}

func TestRunExport_PositionalCharDir(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "book33.yml")
	if got := runExport([]string{testdata.CharDir(), out}, newTextPrinter()); got != 0 {
		t.Errorf("runExport(positional) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "B33S1") {
		t.Errorf("missing B33S1 in output: %s", data)
	}
}

func TestRunExport_AllFlag(t *testing.T) {
	// Two chars in a fake FFXI tree; --all should produce two YAML files.
	src := testdata.CharDir()
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a1b2c3d4", "e5f6a7b8"} {
		charDir := filepath.Join(userDir, id)
		if err := os.Mkdir(charDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// Copy testdata .dat/.ttl files.
		entries, _ := os.ReadDir(src)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, _ := os.ReadFile(filepath.Join(src, e.Name()))
			_ = os.WriteFile(filepath.Join(charDir, e.Name()), data, 0o644)
		}
		// Testdata has no mcr.dat; write a valid empty one so the char dir
		// is discoverable and the parser accepts it.
		_ = os.WriteFile(filepath.Join(charDir, "mcr.dat"), dat.EncodeMacroSet(dat.MacroSet{}), 0o644)
	}

	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	args := []string{"--ffxi-path", ffxiDir, "--all"}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(--all) = %d, want 0", got)
	}
	entries, _ := os.ReadDir(dir)
	var ymlFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yml") {
			ymlFiles = append(ymlFiles, e.Name())
		}
	}
	if len(ymlFiles) != 2 {
		t.Errorf("expected 2 YAML files, got %v", ymlFiles)
	}
}

func TestRunExport_AllWithCharDirErrors(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	charDir := filepath.Join(ffxiDir, "USER", "a1b2c3d4")
	args := []string{"--ffxi-path", ffxiDir, "--all", "--char-dir", charDir}
	if got := runExport(args, newTextPrinter()); got != 1 {
		t.Errorf("runExport(--all --char-dir) = %d, want 1", got)
	}
}

func TestRunExport_AliasAutoPopulatesName(t *testing.T) {
	// Set up a USER dir with a character that has an alias.
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charID := "a1b2c3d4"
	charDir := filepath.Join(userDir, charID)
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Seed with testdata .dat files.
	src := testdata.CharDir()
	entries, _ := os.ReadDir(src)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(src, e.Name()))
		_ = os.WriteFile(filepath.Join(charDir, e.Name()), data, 0o644)
	}
	_ = os.WriteFile(filepath.Join(charDir, "mcr.dat"), dat.EncodeMacroSet(dat.MacroSet{}), 0o644)

	setTestConfig(t, ffxiDir, map[string]string{charID: "Squatched"})

	out := filepath.Join(t.TempDir(), "out.yml")
	args := []string{"--char-dir", charDir, out}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "character: Squatched") {
		t.Errorf("YAML character field missing alias name:\n%s", data)
	}
}

func TestRunExport_CharName(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charID := "a1b2c3d4"
	charDir := filepath.Join(userDir, charID)
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	src := testdata.CharDir()
	entries, _ := os.ReadDir(src)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		srcData, _ := os.ReadFile(filepath.Join(src, e.Name()))
		_ = os.WriteFile(filepath.Join(charDir, e.Name()), srcData, 0o644)
	}
	_ = os.WriteFile(filepath.Join(charDir, "mcr.dat"), dat.EncodeMacroSet(dat.MacroSet{}), 0o644)

	setTestConfig(t, ffxiDir, map[string]string{charID: "Squatched"})

	out := filepath.Join(t.TempDir(), "out.yml")
	args := []string{"--ffxi-path", ffxiDir, "--char-name", "Squatched", out}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(--char-name) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "character: Squatched") {
		t.Errorf("YAML character field missing alias name:\n%s", data)
	}
}

func TestRunExport_WithScopeFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "scoped.yml")
	args := []string{"--char-dir", testdata.CharDir(), "--scope", "B33", out}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(--scope B33) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "level: book") {
		t.Errorf("scoped export missing book scope level: %s", s)
	}
	if !strings.Contains(s, "B33S1") {
		t.Errorf("scoped export missing B33S1: %s", s)
	}
}

func TestRunExport_InvalidScopeFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "bad.yml")
	args := []string{"--char-dir", testdata.CharDir(), "--scope", "B0", out}
	if got := runExport(args, newTextPrinter()); got != 1 {
		t.Errorf("runExport(bad scope) = %d, want 1", got)
	}
}

func TestRunExport_StdoutDefault(t *testing.T) {
	// No output path → write YAML to stdout; command must succeed.
	if got := runExport([]string{"--char-dir", testdata.CharDir()}, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(stdout default) = %d, want 0", got)
	}
}

func TestRunExport_JSON_Success(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yml")
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	args := []string{"--char-dir", testdata.CharDir(), out}
	if got := runExport(args, p); got != 0 {
		t.Fatalf("runExport(JSON) = %d, want 0", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"ok"`) {
		t.Errorf("JSON output missing ok field:\n%s", s)
	}
	if !strings.Contains(s, `"path"`) {
		t.Errorf("JSON output missing path field:\n%s", s)
	}
}

func TestRunExport_AllWithNameError(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	args := []string{"--ffxi-path", ffxiDir, "--all", "--name", "Squatched"}
	if got := runExport(args, newTextPrinter()); got != 1 {
		t.Errorf("runExport(--all --name) = %d, want 1", got)
	}
}

func TestRunExport_AllWithExplicitDir(t *testing.T) {
	ffxiDir := makeExportableFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	outDir := t.TempDir()
	args := []string{"--ffxi-path", ffxiDir, "--all", outDir}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(--all <dir>) = %d, want 0", got)
	}
	entries, _ := os.ReadDir(outDir)
	var ymlFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yml") {
			ymlFiles = append(ymlFiles, e.Name())
		}
	}
	if len(ymlFiles) != 2 {
		t.Errorf("expected 2 YAML files in explicit dir, got %v", ymlFiles)
	}
}

func TestRunExport_AllPathIsFile(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	// Point --all at an existing file, not a directory — should error early.
	f, err := os.CreateTemp(t.TempDir(), "*.yml")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	args := []string{"--ffxi-path", ffxiDir, "--all", f.Name()}
	if got := runExport(args, newTextPrinter()); got != 1 {
		t.Errorf("runExport(--all <file>) = %d, want 1", got)
	}
}

func TestRunExport_AllPathMkdirCreate(t *testing.T) {
	ffxiDir := makeExportableFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	// Point --all at a path that doesn't exist yet — should be created.
	outDir := filepath.Join(t.TempDir(), "new", "subdir")
	args := []string{"--ffxi-path", ffxiDir, "--all", outDir}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(--all <new-dir>) = %d, want 0", got)
	}
	entries, _ := os.ReadDir(outDir)
	var ymlFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yml") {
			ymlFiles = append(ymlFiles, e.Name())
		}
	}
	if len(ymlFiles) != 2 {
		t.Errorf("expected 2 YAML files in created dir, got %v", ymlFiles)
	}
}

func TestRunExport_MultipleScopes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "multi.yml")
	// Two separate --scope flags; both books should appear in selections.
	args := []string{"--char-dir", testdata.CharDir(), "--scope", "B1", "--scope", "B33", out}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(multi scope) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "book: 1") {
		t.Errorf("multi-scope export missing book 1 selection: %s", s)
	}
	if !strings.Contains(s, "book: 33") {
		t.Errorf("multi-scope export missing book 33 selection: %s", s)
	}
}

func TestRunExport_DenseFlag(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "dense.yml")
	args := []string{"--char-dir", testdata.CharDir(), "--dense", "--scope", "B1S1", out}
	if got := runExport(args, newTextPrinter()); got != 0 {
		t.Fatalf("runExport(--dense) = %d, want 0", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	// All 10 ctrl and 10 alt slots of B1S1 should appear.
	if !strings.Contains(s, "ctrl:") || !strings.Contains(s, "alt:") {
		t.Errorf("dense export missing ctrl or alt section:\n%s", s)
	}
	// Empty slots must use comment placeholders, not double-quoted empty strings.
	if !strings.Contains(s, "# Macro Line 1") {
		t.Errorf("dense export should use comment placeholders for empty slots:\n%s", s)
	}
}

// makeExportableFFXITree builds a fake FFXI tree seeded with testdata .dat files
// and a valid mcr.dat, so export can actually read the macro files.
func makeExportableFFXITree(t *testing.T, charIDs ...string) string {
	t.Helper()
	src := testdata.CharDir()
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, id := range charIDs {
		charDir := filepath.Join(userDir, id)
		if err := os.Mkdir(charDir, 0o755); err != nil {
			t.Fatal(err)
		}
		entries, _ := os.ReadDir(src)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, _ := os.ReadFile(filepath.Join(src, e.Name()))
			_ = os.WriteFile(filepath.Join(charDir, e.Name()), data, 0o644)
		}
		_ = os.WriteFile(filepath.Join(charDir, "mcr.dat"), dat.EncodeMacroSet(dat.MacroSet{}), 0o644)
	}
	return ffxiDir
}
