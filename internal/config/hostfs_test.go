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

func TestHostFS_ConfigSplitBrainRegression(t *testing.T) {
	home := t.TempDir()
	wine := &HostFS{GOOS: "windows", UnderWine: true, LinuxHome: home}
	restoreWine := SetHostFSForTest(wine)

	storedCfg, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	openPath, err := OpenPath(storedCfg)
	if err != nil {
		t.Fatal(err)
	}
	if openPath == storedCfg {
		t.Fatal("wine OpenPath must not return POSIX path unchanged (split-brain risk)")
	}
	if !strings.HasPrefix(openPath, `Z:\`) {
		t.Errorf("OpenPath = %q, want Z:\\ prefix", openPath)
	}

	restoreWine()
	restoreLinux := SetHostFSForTest(&HostFS{GOOS: "linux"})
	defer restoreLinux()

	cfg := Empty()
	if err := Save(storedCfg, cfg); err != nil {
		t.Fatal(err)
	}
	hostFile := storedCfg
	if _, err := os.Stat(hostFile); err != nil {
		t.Fatalf("config must exist on host at %q: %v", hostFile, err)
	}
	loaded, err := Load(storedCfg)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != CurrentVersion {
		t.Errorf("loaded version = %d, want %d", loaded.Version, CurrentVersion)
	}
}

func TestHostFS_CrossRuntimeInstallYAML(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "Games", "ffxi")
	ffxiDir := filepath.Join(prefix, "drive_c", "Program Files (x86)", "PlayOnline", "SquareEnix", "FINAL FANTASY XI")
	if err := os.MkdirAll(filepath.Join(ffxiDir, "USER", "abc"), 0o755); err != nil {
		t.Fatal(err)
	}

	wine := &HostFS{GOOS: "windows", UnderWine: true, LinuxHome: home, WinePrefix: prefix}
	stored, err := wine.Stored(`C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI`)
	if err != nil {
		t.Fatal(err)
	}

	restore := SetHostFSForTest(&HostFS{GOOS: "linux"})
	defer restore()
	cfg := Config{
		Version:  CurrentVersion,
		Installs: map[string]Install{"lutris": {Path: stored}},
	}
	cfgPath := filepath.Join(t.TempDir(), "config.yml")
	if err := Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	access, err := ResolveInstallPath(stored)
	if err != nil {
		t.Fatal(err)
	}
	if access != stored {
		t.Errorf("linux ResolveInstallPath = %q, want %q", access, stored)
	}
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Installs["lutris"].Path != stored {
		t.Fatalf("loaded = %q, want %q", loaded.Installs["lutris"].Path, stored)
	}
}
