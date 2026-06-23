package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestRunBackup_Help(t *testing.T) {
	if got := runBackup([]string{"--help"}, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(--help) = %d, want 0", got)
	}
}

func TestRunBackup_ShortHelp(t *testing.T) {
	if got := runBackup([]string{"-h"}, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(-h) = %d, want 0", got)
	}
}

func TestRunBackup_NoArgs(t *testing.T) {
	if got := runBackup(nil, newTextPrinter()); got != 1 {
		t.Errorf("runBackup(nil, newTextPrinter()) = %d, want 1", got)
	}
}

func TestRunBackup_BadCharDir(t *testing.T) {
	if got := runBackup([]string{"/nonexistent/char"}, newTextPrinter()); got != 1 {
		t.Errorf("runBackup(bad dir) = %d, want 1", got)
	}
}

func TestRunBackup_OutAndInPlaceMutuallyExclusive(t *testing.T) {
	tmp := prepBackupCharDir(t)
	out := t.TempDir()
	args := []string{"--char", tmp, "--out", out, "--in-place"}
	if got := runBackup(args, newTextPrinter()); got != 1 {
		t.Errorf("runBackup(--out + --in-place) = %d, want 1", got)
	}
}

func TestRunBackup_DefaultDestIsCWD(t *testing.T) {
	tmp := prepBackupCharDir(t)
	cwd, _ := os.Getwd()
	if got := runBackup([]string{"--char", tmp}, newTextPrinter()); got != 0 {
		t.Fatalf("runBackup(default dest) = %d, want 0", got)
	}
	assertBackupUnder(t, cwd, filepath.Base(tmp))
	// Clean up so we don't leave files in the test working directory.
	entries, _ := os.ReadDir(cwd)
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), filepath.Base(tmp)+"_") {
			_ = os.RemoveAll(filepath.Join(cwd, e.Name()))
		}
	}
}

func TestRunBackup_Success_Positional(t *testing.T) {
	tmp := prepBackupCharDir(t)
	dest := t.TempDir()
	if got := runBackup([]string{"--out", dest, tmp}, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(positional) = %d, want 0", got)
	}
	assertBackupUnder(t, dest, filepath.Base(tmp))
}

func TestRunBackup_Success_CharFlag(t *testing.T) {
	tmp := prepBackupCharDir(t)
	dest := t.TempDir()
	if got := runBackup([]string{"--char", tmp, "--out", dest}, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(--char) = %d, want 0", got)
	}
	assertBackupUnder(t, dest, filepath.Base(tmp))
}

func TestRunBackup_InPlace(t *testing.T) {
	tmp := prepBackupCharDir(t)
	if got := runBackup([]string{"--char", tmp, "--in-place"}, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(--in-place) = %d, want 0", got)
	}
	assertBackupUnder(t, filepath.Join(tmp, "backups"), filepath.Base(tmp))
}

// assertBackupUnder verifies that a backup directory for charID was created
// under parent and contains at least one .dat file.
func assertBackupUnder(t *testing.T, parent, charID string) {
	t.Helper()
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", parent, err)
	}
	var stamped string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), charID+"_") {
			stamped = filepath.Join(parent, e.Name())
			break
		}
	}
	if stamped == "" {
		t.Fatalf("no backup directory matching %s_* found under %s", charID, parent)
	}
	files, _ := os.ReadDir(stamped)
	var hasDat bool
	for _, f := range files {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".dat") {
			hasDat = true
			break
		}
	}
	if !hasDat {
		t.Errorf("backup at %s contains no .dat files", stamped)
	}
}

func TestRunBackup_All(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	dest := t.TempDir()
	args := []string{"--ffxi-path", ffxiDir, "--all", "--out", dest}
	if got := runBackup(args, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(--all) = %d, want 0", got)
	}
	assertBackupUnder(t, dest, "a1b2c3d4")
	assertBackupUnder(t, dest, "e5f6a7b8")
}

func TestRunBackup_AllInPlace(t *testing.T) {
	ffxiDir, userDir, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	_ = userDir
	args := []string{"--ffxi-path", ffxiDir, "--all", "--in-place"}
	if got := runBackup(args, newTextPrinter()); got != 0 {
		t.Errorf("runBackup(--all --in-place) = %d, want 0", got)
	}
	assertBackupUnder(t, filepath.Join(ffxiDir, "USER", "a1b2c3d4", "backups"), "a1b2c3d4")
	assertBackupUnder(t, filepath.Join(ffxiDir, "USER", "e5f6a7b8", "backups"), "e5f6a7b8")
}

func prepBackupCharDir(t *testing.T) string {
	t.Helper()
	src := testdata.CharDir()
	tmp := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", src, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(tmp, e.Name()), data, 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", e.Name(), err)
		}
	}
	return tmp
}
