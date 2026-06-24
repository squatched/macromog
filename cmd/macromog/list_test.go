package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
)

// blankDatBytes returns a valid-header, all-zero macro set (no macros).
func blankDatBytes() []byte {
	buf := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(buf[0:4], dat.MagicVersion)
	return buf
}

// namedDatBytes returns a dat file where ctrl[0] has the given macro name.
func namedDatBytes(name string) []byte {
	buf := blankDatBytes()
	nameOff := dat.HeaderSize + dat.MacroPrefixSize + dat.LineCount*dat.LineSize
	copy(buf[nameOff:nameOff+dat.NameSize], []byte(name))
	return buf
}

// makeTestUserDir creates a synthetic USER directory for CLI tests.
func makeTestUserDir(t *testing.T) string {
	t.Helper()
	userDir := t.TempDir()

	// charA: one non-empty book (book 1)
	charA := filepath.Join(userDir, "aabbcc")
	if err := os.MkdirAll(charA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(charA, "mcr.dat"), namedDatBytes("Cure"), 0o644); err != nil {
		t.Fatal(err)
	}

	// charB: no macros (blank dat)
	charB := filepath.Join(userDir, "ddeeff")
	if err := os.MkdirAll(charB, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(charB, "mcr.dat"), blankDatBytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	return userDir
}

func TestRunList_Help(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		if got := runList([]string{flag}, newTextPrinter()); got != 0 {
			t.Errorf("runList(%s) = %d, want 0", flag, got)
		}
	}
}

func TestRunList_BadFlag(t *testing.T) {
	if got := runList([]string{"--unknown-flag"}, newTextPrinter()); got != 1 {
		t.Errorf("runList(bad flag) = %d, want 1", got)
	}
}

func TestRunList_MissingFFXIPath(t *testing.T) {
	// Supply a path that does not contain a USER subdir.
	if got := runList([]string{"--ffxi-path", t.TempDir()}, newTextPrinter()); got != 1 {
		t.Errorf("runList(missing USER dir) = %d, want 1", got)
	}
}

func TestRunList_WithFFXIPath(t *testing.T) {
	userDir := makeTestUserDir(t)
	// ffxi-path points to the parent of USER.
	ffxiRoot := filepath.Dir(userDir)
	// Rename temp USER dir so --ffxi-path sees it as "USER".
	renamedUser := filepath.Join(ffxiRoot, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}

	if got := runList([]string{"--ffxi-path", ffxiRoot}, newTextPrinter()); got != 0 {
		t.Errorf("runList(--ffxi-path) = %d, want 0", got)
	}
}

func TestRunList_CharDir_NoMacros(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), blankDatBytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := runList([]string{"--char-dir", dir}, newTextPrinter()); got != 0 {
		t.Errorf("runList(--char-dir empty) = %d, want 0", got)
	}
}

func TestRunList_CharDir_WithMacros(t *testing.T) {
	// Use the existing test fixture which has macros in book 33.
	if got := runList([]string{"--char-dir", testdata.CharDir()}, newTextPrinter()); got != 0 {
		t.Errorf("runList(--char-dir testdata) = %d, want 0", got)
	}
}

func TestRunList_CharDir_NotFound(t *testing.T) {
	if got := runList([]string{"--char-dir", "/nonexistent/char"}, newTextPrinter()); got != 1 {
		t.Errorf("runList(--char-dir missing) = %d, want 1", got)
	}
}

func TestRunList_CharDir_ShowsBookIndex(t *testing.T) {
	// Book 33 set 1 is present in the testdata fixture.
	// We just verify the exit code and that the logic doesn't crash.
	if got := run([]string{"macromog", "list", "--char-dir", testdata.CharDir()}); got != 0 {
		t.Errorf("run(list --char-dir testdata) = %d, want 0", got)
	}
}

func TestRunList_UserDir_MultipleChars(t *testing.T) {
	userDir := makeTestUserDir(t)
	// Point --ffxi-path to parent, with "USER" subdir = userDir.
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}

	if got := runList([]string{"--ffxi-path", parent}, newTextPrinter()); got != 0 {
		t.Errorf("runList(multi-char) = %d, want 0", got)
	}
}

func TestRunList_EmptyUserDir(t *testing.T) {
	parent := t.TempDir()
	userDir := filepath.Join(parent, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := runList([]string{"--ffxi-path", parent}, newTextPrinter()); got != 0 {
		t.Errorf("runList(empty USER) = %d, want 0", got)
	}
}

func TestRunList_CharName(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}

	setTestConfig(t, parent, map[string]string{"aabbcc": "Squatched"})

	if got := runList([]string{"--ffxi-path", parent, "--char-name", "Squatched"}, newTextPrinter()); got != 0 {
		t.Errorf("runList(--char-name) = %d, want 0", got)
	}
}

func TestRunList_CharName_NotFound(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}

	if got := runList([]string{"--ffxi-path", parent, "--char-name", "Nobody"}, newTextPrinter()); got != 1 {
		t.Errorf("runList(--char-name unknown) = %d, want 1", got)
	}
}

func TestRunList_CharName_DirMissing(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}
	setTestConfig(t, parent, map[string]string{"aabbcc": "Squatched"})
	// Delete the directory the alias points to.
	if err := os.RemoveAll(filepath.Join(renamedUser, "aabbcc")); err != nil {
		t.Fatal(err)
	}

	if got := runList([]string{"--ffxi-path", parent, "--char-name", "Squatched"}, newTextPrinter()); got != 1 {
		t.Errorf("runList(--char-name deleted dir) = %d, want 1", got)
	}
}

func TestRunList_ShowsAliasInOutput(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}
	setTestConfig(t, parent, map[string]string{"aabbcc": "Squatched"})

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	if got := runList([]string{"--ffxi-path", parent}, p); got != 0 {
		t.Fatalf("runList = %d, want 0", got)
	}
	if !strings.Contains(buf.String(), "Squatched") {
		t.Errorf("output missing alias name Squatched:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "aabbcc") {
		t.Errorf("output missing hex ID aabbcc:\n%s", buf.String())
	}
}

func TestRunList_ShowsAliasForSingleChar(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}
	setTestConfig(t, parent, map[string]string{"aabbcc": "Squatched"})

	charDir := filepath.Join(renamedUser, "aabbcc")
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	if got := runList([]string{"--char-dir", charDir}, p); got != 0 {
		t.Fatalf("runList(--char-dir) = %d, want 0", got)
	}
	if !strings.Contains(buf.String(), "Squatched") {
		t.Errorf("single-char output missing alias name:\n%s", buf.String())
	}
}

func TestRunList_InvalidConfig(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("version: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)

	if got := runList([]string{"--ffxi-path", parent}, newTextPrinter()); got != 1 {
		t.Errorf("runList(invalid config) = %d, want 1", got)
	}
}

func TestRunList_WideCharAlignment(t *testing.T) {
	// A Japanese alias (wide chars, 2 display cols each) and a plain ASCII ID
	// must produce "(no macros)" at the same terminal display column.
	parent := t.TempDir()
	userDir := filepath.Join(parent, "USER")
	for _, id := range []string{"aabbcc", "ddeeff"} {
		dir := filepath.Join(userDir, id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), blankDatBytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// 玄白侍士: 4 wide chars = 8 display columns
	setTestConfig(t, parent, map[string]string{"aabbcc": "玄白侍士"})

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	if got := runList([]string{"--ffxi-path", parent}, p); got != 0 {
		t.Fatalf("runList = %d, want 0", got)
	}

	// Collect the display-column position of "(no macros)" on each character line.
	var cols []int
	for _, line := range strings.Split(buf.String(), "\n") {
		if idx := strings.Index(line, "(no macros)"); idx != -1 {
			cols = append(cols, visibleWidth(line[:idx]))
		}
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 lines with '(no macros)', got %d:\n%s", len(cols), buf.String())
	}
	if cols[0] != cols[1] {
		t.Errorf("'(no macros)' at display cols %v — wide chars broke alignment:\n%s", cols, buf.String())
	}
}

func TestRunList_JSON_All(t *testing.T) {
	userDir := makeTestUserDir(t)
	parent := t.TempDir()
	renamedUser := filepath.Join(parent, "USER")
	if err := os.Rename(userDir, renamedUser); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runList([]string{"--ffxi-path", parent}, p); got != 0 {
		t.Fatalf("runList(JSON all) = %d, want 0", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"user_dir"`) {
		t.Errorf("JSON output missing user_dir field:\n%s", s)
	}
	if !strings.Contains(s, `"characters"`) {
		t.Errorf("JSON output missing characters field:\n%s", s)
	}
}

func TestRunList_JSON_SingleChar(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), blankDatBytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runList([]string{"--char-dir", dir}, p); got != 0 {
		t.Fatalf("runList(JSON single) = %d, want 0", got)
	}
	s := buf.String()
	if !strings.Contains(s, `"character"`) {
		t.Errorf("JSON output missing character field:\n%s", s)
	}
	if !strings.Contains(s, `"books"`) {
		t.Errorf("JSON output missing books field:\n%s", s)
	}
}

func TestRunList_CharDir_NotADir(t *testing.T) {
	// --char-dir pointing at a file rather than a directory must fail.
	f, err := os.CreateTemp(t.TempDir(), "notadir")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if got := runList([]string{"--char-dir", f.Name()}, newTextPrinter()); got != 1 {
		t.Errorf("runList(--char-dir file) = %d, want 1", got)
	}
}

func TestRunList_CharDir_BookName(t *testing.T) {
	dir := t.TempDir()

	// Write book title "WHM75NIN" into mcr.ttl (20 slots × 16 bytes).
	ttl := make([]byte, dat.HeaderSize+20*dat.BookNameSize)
	binary.LittleEndian.PutUint32(ttl[0:4], dat.MagicVersion)
	copy(ttl[dat.HeaderSize:dat.HeaderSize+dat.BookNameSize], []byte("WHM75NIN"))
	if err := os.WriteFile(filepath.Join(dir, "mcr.ttl"), ttl, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mcr.dat"), namedDatBytes("Cure"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	code := runList([]string{"--char-dir", dir}, p)

	if code != 0 {
		t.Fatalf("runList = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "WHM75NIN") {
		t.Errorf("output missing book name WHM75NIN:\n%s", buf.String())
	}
}
