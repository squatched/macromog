package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

func TestRegisterInstall_WineStoredPathOnHost(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "Games", "ffxi")
	ffxiDir := filepath.Join(prefix, "drive_c", "Program Files (x86)", "PlayOnline", "SquareEnix", "FINAL FANTASY XI")
	charDir := filepath.Join(ffxiDir, "USER", "a1b2c3d4")
	if err := os.MkdirAll(charDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(charDir, "mcr.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	wine := &config.HostFS{
		GOOS:       "windows",
		UnderWine:  true,
		LinuxHome:  home,
		WinePrefix: prefix,
	}
	wineInput := `C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI`
	stored, err := wine.Stored(wineInput)
	if err != nil {
		t.Fatal(err)
	}
	if stored != ffxiDir {
		t.Fatalf("Stored() = %q, want %q", stored, ffxiDir)
	}
	if strings.Contains(stored, `:`) {
		t.Fatalf("stored path must be POSIX, got %q", stored)
	}

	cfgPath := filepath.Join(home, ".config", "macromog", "config.yml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(cfgPath, config.Empty()); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", cfgPath)

	restore := config.SetHostFSForTest(&config.HostFS{GOOS: "linux"})
	defer restore()

	session, err := openConfig()
	if err != nil {
		t.Fatal(err)
	}
	if err := registerInstall(session, "lutris", stored, registerInstallOpts{}); err != nil {
		t.Fatal(err)
	}
	if session.cfg.Installs["lutris"].Path != stored {
		t.Errorf("cfg path = %q, want %q", session.cfg.Installs["lutris"].Path, stored)
	}
}

func TestRegisterInstall_DuplicateInstall(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	session, err := openConfig()
	if err != nil {
		t.Fatal(err)
	}
	if err := registerInstall(session, "steam", ffxiDir, registerInstallOpts{}); err != nil {
		t.Fatal(err)
	}
	err = registerInstall(session, "steam", ffxiDir, registerInstallOpts{})
	if err == nil {
		t.Fatal("expected duplicate install error")
	}
}

func TestRegisterInstall_MissingUSER(t *testing.T) {
	resetTestConfig(t)
	session, err := openConfig()
	if err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(t.TempDir(), "no-user-here")
	err = registerInstall(session, "steam", bad, registerInstallOpts{})
	if err == nil {
		t.Fatal("expected missing USER error")
	}
}
