package config

import (
	"strings"
	"testing"
)

func TestIsDrivePath_Edges(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{in: `C:\foo`, want: true},
		{in: `c:\foo`, want: true},
		{in: `C:`, want: false},
		{in: `C:relative`, want: false},
		{in: `1:\bad`, want: false},
		{in: `/posix`, want: false},
	}
	for _, tc := range tests {
		if got := isDrivePath(tc.in); got != tc.want {
			t.Errorf("isDrivePath(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestCanonicalZDrivePath_LowercaseZ(t *testing.T) {
	got, err := canonicalZDrivePath(`z:\home\adventurer\pfx\drive_c\FFXI`)
	if err != nil {
		t.Fatal(err)
	}
	want := "/home/adventurer/pfx/drive_c/FFXI"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToStoredPath_TrailingSlashHome(t *testing.T) {
	got, err := toStoredPath("/home/squatched/Games/ffxi/")
	if err != nil {
		t.Fatal(err)
	}
	want := "/home/squatched/Games/ffxi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValidate_RejectsWindowsStyleInstallPath(t *testing.T) {
	cfg := Config{
		Version: 1,
		Installs: map[string]Install{
			"steam": {Path: `C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI`},
		},
	}
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("expected validation error for C:\\ path in yaml")
	}
	if !strings.Contains(err.Error(), "path must be absolute and normalized") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_RejectsTrailingSlashStoredPath(t *testing.T) {
	dir := t.TempDir()
	withSlash := dir + "/"
	cfg := Config{
		Version: 1,
		Installs: map[string]Install{
			"steam": {Path: withSlash},
		},
	}
	err := Validate(&cfg)
	if err == nil {
		t.Fatal("expected validation error for trailing slash path")
	}
}

func TestOpenPath_WineHostFS_UsesZNotPOSIX(t *testing.T) {
	home := t.TempDir()
	wine := &HostFS{
		GOOS:      "windows",
		UnderWine: true,
		LinuxHome: home,
	}
	restore := SetHostFSForTest(wine)
	defer restore()

	stored := hostpath(home, ".config", "macromog", "config.yml")
	got, err := OpenPath(stored)
	if err != nil {
		t.Fatal(err)
	}
	if got == stored {
		t.Fatalf("OpenPath returned POSIX %q; want Z: mapping under wine", got)
	}
	if !strings.HasPrefix(got, `Z:\`) {
		t.Errorf("OpenPath = %q, want Z:\\ prefix", got)
	}
}

func TestHostFS_StoredProtonLayout(t *testing.T) {
	home := "/home/adventurer"
	prefix := hostpath(home, ".steam", "steam", "steamapps", "compatdata", "230330", "pfx")
	wine := &HostFS{
		GOOS:       "windows",
		UnderWine:  true,
		LinuxHome:  home,
		WinePrefix: prefix,
	}
	input := `Z:\home\adventurer\.steam\steam\steamapps\compatdata\230330\pfx\drive_c\FFXI`
	stored, err := wine.Stored(input)
	if err != nil {
		t.Fatal(err)
	}
	want := "/home/adventurer/.steam/steam/steamapps/compatdata/230330/pfx/drive_c/FFXI"
	if stored != want {
		t.Errorf("got %q, want %q", stored, want)
	}
}

func TestResolveForWine_RelativeFallsBackToNormalizePath(t *testing.T) {
	got, err := resolveForWine("relative\\path")
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("expected non-empty normalized path")
	}
}

func TestHostFS_ConfigPath_WindowsGOOS(t *testing.T) {
	t.Setenv("WINE_HOST_XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	wine := &HostFS{
		GOOS:      "windows",
		UnderWine: true,
		LinuxHome: home,
	}
	restore := SetHostFSForTest(wine)
	defer restore()

	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	want := hostpath(home, ".config", "macromog", "config.yml")
	if got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestHostFS_MACROMOG_CONFIG_AccessUnderWine(t *testing.T) {
	t.Setenv("MACROMOG_CONFIG", "/home/squatched/.config/macromog/config.yml")
	wine := &HostFS{GOOS: "windows", UnderWine: true, LinuxHome: "/home/squatched"}
	restore := SetHostFSForTest(wine)
	defer restore()

	got, err := OpenPath("/home/squatched/.config/macromog/config.yml")
	if err != nil {
		t.Fatal(err)
	}
	want := `Z:\home\squatched\.config\macromog\config.yml`
	if got != want {
		t.Errorf("OpenPath = %q, want %q", got, want)
	}
}
