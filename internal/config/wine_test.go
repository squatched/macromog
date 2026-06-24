package config

import (
	"path/filepath"
	"testing"
)

func TestCanonicalForWine_DrivePath(t *testing.T) {
	home := "/home/adventurer"
	prefix := filepath.Join(home, ".wine")
	t.Setenv("WINEPREFIX", prefix)

	got, err := canonicalForWine(`C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI`)
	if err != nil {
		t.Fatal(err)
	}
	want := "/home/adventurer/.wine/drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCanonicalForWine_ZDrive(t *testing.T) {
	got, err := canonicalForWine(`Z:\home\adventurer\.steam\steamapps\compatdata\230330\pfx\drive_c\FFXI`)
	if err != nil {
		t.Fatal(err)
	}
	want := "/home/adventurer/.steam/steamapps/compatdata/230330/pfx/drive_c/FFXI"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCanonicalForWine_PosixPassthrough(t *testing.T) {
	in := "/home/adventurer/Games/ffxi/drive_c/Program Files (x86)/FINAL FANTASY XI"
	got, err := canonicalForWine(in)
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Errorf("got %q, want %q", got, in)
	}
}

func TestResolveForWine_PosixToZ(t *testing.T) {
	stored := "/home/adventurer/.wine/drive_c/Program Files (x86)/FINAL FANTASY XI"
	got, err := resolveForWine(stored)
	if err != nil {
		t.Fatal(err)
	}
	want := `Z:\home\adventurer\.wine\drive_c\Program Files (x86)\FINAL FANTASY XI`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunningUnderWine_Linux(t *testing.T) {
	if RunningUnderWine() {
		t.Error("expected false on linux test runner")
	}
}

func TestCanonicalInstallPath_LinuxNative(t *testing.T) {
	dir := t.TempDir()
	got, err := CanonicalInstallPath(dir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := NormalizePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}