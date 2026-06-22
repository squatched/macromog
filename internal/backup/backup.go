package backup

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Backup copies all *.dat and *.ttl files from charDir into a new subdirectory
// of destDir named "<charID>_YYYYMMDD_HHMMSS" and returns that subdirectory's path.
func Backup(charDir, destDir string) (string, error) {
	charID := filepath.Base(charDir)
	stamp := time.Now().UTC().Format("20060102_150405")
	backupDir := filepath.Join(destDir, charID+"_"+stamp)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(charDir)
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
		if err := copyFile(filepath.Join(charDir, e.Name()), filepath.Join(backupDir, e.Name())); err != nil {
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
