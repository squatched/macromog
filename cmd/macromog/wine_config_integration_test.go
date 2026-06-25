package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

// TestIntegration_WineStoredPath_LinuxHost verifies that a POSIX path written
// the way Wine-side add-install would store it is usable by the native Linux CLI.
func TestIntegration_WineStoredPath_LinuxHost(t *testing.T) {
	root := t.TempDir()
	storedPath := filepath.Join(
		root,
		".wine", "drive_c",
		"Program Files (x86)", "PlayOnline", "SquareEnix", "FINAL FANTASY XI",
	)
	charID := "a1b2c3d4"
	for _, d := range []string{
		storedPath,
		filepath.Join(storedPath, "USER"),
		filepath.Join(storedPath, "USER", charID),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(storedPath, "USER", charID, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	normStored, err := config.NormalizePath(storedPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Version: 1,
		Installs: map[string]config.Install{
			"wine": {Path: normStored},
		},
	}
	session := testSession(t, cfg)

	var listBuf bytes.Buffer
	if got := runList([]string{"--install", "wine"}, NewPrinter(&listBuf, FormatText)); got != 0 {
		t.Fatalf("list = %d", got)
	}
	if !strings.Contains(listBuf.String(), charID) {
		t.Errorf("list output missing character dir:\n%s", listBuf.String())
	}

	name, inst, err := config.FindInstallByPath(&session.cfg, normStored)
	if err != nil {
		t.Fatal(err)
	}
	if name != "wine" || inst == nil {
		t.Fatalf("FindInstallByPath = %q, %v", name, inst)
	}

	access, err := config.ResolveInstallPath(inst.Path)
	if err != nil {
		t.Fatal(err)
	}
	if access != normStored {
		t.Errorf("ResolveInstallPath = %q, want %q", access, normStored)
	}
}
