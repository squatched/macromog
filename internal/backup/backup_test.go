package backup_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/backup"
	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestBackup_CreatesTimestampedDirectory(t *testing.T) {
	tmp := prepCharDir(t)

	dir, err := backup.Backup(tmp)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		t.Errorf("backup dir not created: %s", dir)
	}
	if !strings.HasPrefix(dir, filepath.Join(tmp, "backups")+string(filepath.Separator)) {
		t.Errorf("backup dir %s not under <char>/backups/", dir)
	}
}

func TestBackup_CopiesDatAndTtlFiles(t *testing.T) {
	src := testdata.CharDir()
	tmp := prepCharDir(t)

	backupDir, err := backup.Backup(tmp)
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
	// Plant a file with a non-dat/ttl extension and a subdirectory
	_ = os.WriteFile(filepath.Join(tmp, "notes.txt"), []byte("ignore me"), 0o644)
	_ = os.MkdirAll(filepath.Join(tmp, "subdir"), 0o755)

	backupDir, err := backup.Backup(tmp)
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

	dir, err := backup.Backup(tmp)
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
