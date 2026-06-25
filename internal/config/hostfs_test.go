package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostFS_LutrisInstallRoundTrip(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "Games", "final-fantasy-xi-online")
	userDir := filepath.Join(prefix, ffxiUserSuffix)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wine := &HostFS{
		GOOS:       "windows",
		UnderWine:  true,
		LinuxHome:  home,
		WinePrefix: prefix,
	}

	original := `C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI`
	stored, err := wine.Stored(original)
	if err != nil {
		t.Fatal(err)
	}
	wantStored := filepath.Join(prefix, "drive_c", "Program Files (x86)", "PlayOnline", "SquareEnix", "FINAL FANTASY XI")
	if stored != wantStored {
		t.Fatalf("Stored() = %q, want %q", stored, wantStored)
	}

	wineAccess, err := wine.Access(stored)
	if err != nil {
		t.Fatal(err)
	}
	wantZ := `Z:\` + strings.ReplaceAll(strings.TrimPrefix(stored, "/"), "/", `\`)
	if wineAccess != wantZ {
		t.Errorf("wine Access() = %q, want %q", wineAccess, wantZ)
	}

	linux := &HostFS{GOOS: "linux"}
	linuxAccess, err := linux.Access(stored)
	if err != nil {
		t.Fatal(err)
	}
	if linuxAccess != stored {
		t.Errorf("linux Access() = %q, want %q", linuxAccess, stored)
	}

	cfgDir := filepath.Join(home, ".config", "macromog")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(cfgDir, "config.yml")
	cfg := Config{
		Version: CurrentVersion,
		Installs: map[string]Install{
			"lutris": {Path: stored},
		},
	}
	data, err := MarshalYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Installs["lutris"].Path != stored {
		t.Fatalf("loaded path = %q, want %q", loaded.Installs["lutris"].Path, stored)
	}
}

func TestHostFS_SetForTest(t *testing.T) {
	restore := SetHostFSForTest(&HostFS{GOOS: "linux", UnderWine: true})
	defer restore()
	if !RunningUnderWine() {
		t.Fatal("expected RunningUnderWine true after SetHostFSForTest")
	}
}
