package backup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/backup"
	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestBackup_CreatesNamedDirectory(t *testing.T) {
	tmp := prepCharDir(t)
	dest := t.TempDir()

	dir, err := backup.Backup(tmp, dest, "")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		t.Errorf("backup dir not created: %s", dir)
	}

	// Must be directly under dest (no extra nesting).
	if filepath.Dir(dir) != dest {
		t.Errorf("backup dir %s not directly under dest %s", dir, dest)
	}

	// Without charName: name is "<charID>_backup_<stamp>".
	charID := filepath.Base(tmp)
	base := filepath.Base(dir)
	if !strings.HasPrefix(base, charID+"_backup_") {
		t.Errorf("backup dir name %q does not match %q", base, charID+"_backup_<stamp>")
	}
}

func TestBackup_CreatesNamedDirectory_WithCharName(t *testing.T) {
	tmp := prepCharDir(t)
	dest := t.TempDir()

	dir, err := backup.Backup(tmp, dest, "Squatched")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	charID := filepath.Base(tmp)
	base := filepath.Base(dir)
	wantPrefix := "Squatched_" + charID + "_backup_"
	if !strings.HasPrefix(base, wantPrefix) {
		t.Errorf("backup dir name %q does not match %q", base, wantPrefix+"<stamp>")
	}
}

func TestBackup_CopiesDatAndTtlFiles(t *testing.T) {
	src := testdata.CharDir()
	tmp := prepCharDir(t)
	dest := t.TempDir()

	backupDir, err := backup.Backup(tmp, dest, "")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	entries, _ := os.ReadDir(src)
	for _, e := range entries {
		lower := strings.ToLower(e.Name())
		if strings.HasSuffix(lower, ".dat") || strings.HasSuffix(lower, ".ttl") {
			if _, err := os.Stat(filepath.Join(backupDir, e.Name())); err != nil {
				t.Errorf("backup missing %s: %v", e.Name(), err)
			}
		}
	}
}

func TestBackup_SkipsSubdirectoriesAndOtherFiles(t *testing.T) {
	tmp := prepCharDir(t)
	dest := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "notes.txt"), []byte("ignore me"), 0o644)
	_ = os.MkdirAll(filepath.Join(tmp, "subdir"), 0o755)

	backupDir, err := backup.Backup(tmp, dest, "")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	entries, _ := os.ReadDir(backupDir)
	for _, e := range entries {
		if e.IsDir() {
			t.Errorf("backup contains unexpected subdirectory %s", e.Name())
		}
		lower := strings.ToLower(e.Name())
		if !strings.HasSuffix(lower, ".dat") && !strings.HasSuffix(lower, ".ttl") {
			t.Errorf("backup contains unexpected file %s", e.Name())
		}
	}
}

func TestBackup_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	dest := t.TempDir()

	dir, err := backup.Backup(tmp, dest, "")
	if err != nil {
		t.Fatalf("Backup empty dir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("backup dir not created for empty char dir: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("expected 0 files in backup of empty dir, got %d", len(entries))
	}
}

func TestBackup_UnreadableSourceFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses file permissions; skipping unreadable-file test")
	}
	tmp := t.TempDir()
	datPath := filepath.Join(tmp, "mcr.dat")
	if err := os.WriteFile(datPath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(datPath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(datPath, 0o644) })

	dest := t.TempDir()
	if _, err := backup.Backup(tmp, dest, ""); err == nil {
		t.Error("expected error when source file is unreadable")
	}
}

// prepCharDir copies the test fixture into a fresh temp dir so tests can write
// into it without touching the checked-in testdata.
func prepCharDir(t *testing.T) string {
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
