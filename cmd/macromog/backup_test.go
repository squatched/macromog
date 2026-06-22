package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestRunBackup_Help(t *testing.T) {
	if got := runBackup([]string{"--help"}); got != 0 {
		t.Errorf("runBackup(--help) = %d, want 0", got)
	}
}

func TestRunBackup_ShortHelp(t *testing.T) {
	if got := runBackup([]string{"-h"}); got != 0 {
		t.Errorf("runBackup(-h) = %d, want 0", got)
	}
}

func TestRunBackup_NoArgs(t *testing.T) {
	if got := runBackup(nil); got != 1 {
		t.Errorf("runBackup(nil) = %d, want 1", got)
	}
}

func TestRunBackup_BadCharDir(t *testing.T) {
	if got := runBackup([]string{"/nonexistent/char"}); got != 1 {
		t.Errorf("runBackup(bad dir) = %d, want 1", got)
	}
}

func TestRunBackup_Success_Positional(t *testing.T) {
	tmp := prepBackupCharDir(t)
	if got := runBackup([]string{tmp}); got != 0 {
		t.Errorf("runBackup(positional) = %d, want 0", got)
	}
	assertBackupCreated(t, tmp)
}

func TestRunBackup_Success_CharFlag(t *testing.T) {
	tmp := prepBackupCharDir(t)
	if got := runBackup([]string{"--char", tmp}); got != 0 {
		t.Errorf("runBackup(--char) = %d, want 0", got)
	}
	assertBackupCreated(t, tmp)
}

func assertBackupCreated(t *testing.T, charDir string) {
	t.Helper()
	backupsDir := filepath.Join(charDir, "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("backups dir missing: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no backup subdirectory created")
	}
	stamped := filepath.Join(backupsDir, entries[0].Name())
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
