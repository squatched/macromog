package backup

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Backup copies all *.dat and *.ttl files in dir to a timestamped
// subdirectory and returns its path.
func Backup(dir string) (string, error) {
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
