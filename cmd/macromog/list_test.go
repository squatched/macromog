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
