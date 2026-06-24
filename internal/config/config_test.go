package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func absPath(t *testing.T, elems ...string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(elems...))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if len(cfg.Installs) != 0 {
		t.Errorf("installs = %v, want nil/empty", cfg.Installs)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(":\tnot yaml"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parsing config") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_InvalidSemantics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	content := "version: 1\ndefault_install: missing\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	ffxi := absPath(t, dir, "ffxi")
	offering := false
	cfg := Config{
		Version:        1,
		DefaultInstall: "steam",
		Preferences:    &Preferences{DefaultOffering: &offering},
		Installs: map[string]Install{
			"steam": {
				Path:             ffxi,
				DefaultCharacter: "a1b2c3d4",
				Characters: map[string]Character{
					"a1b2c3d4": {Name: "Squatched"},
				},
			},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.DefaultInstall != "steam" {
		t.Errorf("default_install = %q, want steam", loaded.DefaultInstall)
	}
	if got := loaded.Installs["steam"].Characters["a1b2c3d4"].Name; got != "Squatched" {
		t.Errorf("alias = %q, want Squatched", got)
	}
	if got := loaded.Preferences.DefaultOffering; got == nil || *got != false {
		t.Errorf("default_offering = %v, want false", got)
	}
}

func TestSave_RejectsInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	cfg := Config{Version: 99}
	if err := Save(path, cfg); err == nil {
		t.Error("Save should reject invalid config")
	}
}

func TestSave_AtomicReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := Save(path, Empty()); err != nil {
		t.Fatal(err)
	}
	info1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	ffxi := absPath(t, dir, "ffxi")
	cfg := Config{
		Version: 1,
		Installs: map[string]Install{
			"steam": {Path: ffxi},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	info2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "steam") {
		t.Errorf("expected updated content, got: %s", data)
	}
	if info1.ModTime().After(info2.ModTime()) {
		t.Error("expected file to be replaced in place")
	}
}

func TestEnsure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.yml")

	if err := Ensure(path); err != nil {
		t.Fatalf("first Ensure: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "version: 1") {
		t.Errorf("unexpected content: %s", data)
	}

	if err := Ensure(path); err != nil {
		t.Fatalf("second Ensure: %v", err)
	}
}

func TestValidate(t *testing.T) {
	validPath := absPath(t, t.TempDir(), "ffxi")

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "version zero defaults to one",
			cfg:  Config{Version: 0},
		},
		{
			name:    "unsupported version",
			cfg:     Config{Version: 2},
			wantErr: "unsupported config version",
		},
		{
			name: "default_install unknown",
			cfg: Config{
				Version:        1,
				DefaultInstall: "ghost",
				Installs:       map[string]Install{"steam": {Path: validPath}},
			},
			wantErr: `default_install "ghost" references unknown install`,
		},
		{
			name: "default_install with nil installs",
			cfg: Config{
				Version:        1,
				DefaultInstall: "steam",
			},
			wantErr: `default_install "steam" references unknown install`,
		},
		{
			name: "empty install path",
			cfg: Config{
				Version:  1,
				Installs: map[string]Install{"steam": {Path: ""}},
			},
			wantErr: `install "steam": path must not be empty`,
		},
		{
			name: "relative install path",
			cfg: Config{
				Version:  1,
				Installs: map[string]Install{"steam": {Path: "relative/path"}},
			},
			wantErr: "path must be absolute and normalized",
		},
		{
			name: "default_character missing from characters",
			cfg: Config{
				Version: 1,
				Installs: map[string]Install{
					"steam": {Path: validPath, DefaultCharacter: "a1b2c3d4"},
				},
			},
			wantErr: `default_character "a1b2c3d4" is not configured`,
		},
		{
			name: "empty character id",
			cfg: Config{
				Version: 1,
				Installs: map[string]Install{
					"steam": {
						Path:       validPath,
						Characters: map[string]Character{"": {Name: "Bad"}},
					},
				},
			},
			wantErr: "character id must not be empty",
		},
		{
			name: "empty alias name",
			cfg: Config{
				Version: 1,
				Installs: map[string]Install{
					"steam": {
						Path:       validPath,
						Characters: map[string]Character{"a1": {Name: "  "}},
					},
				},
			},
			wantErr: `alias for "a1" must not be empty`,
		},
		{
			name: "duplicate alias case insensitive",
			cfg: Config{
				Version: 1,
				Installs: map[string]Install{
					"steam": {
						Path: validPath,
						Characters: map[string]Character{
							"a1": {Name: "Squatched"},
							"b2": {Name: "SQUATCHED"},
						},
					},
				},
			},
			wantErr: `duplicate alias "SQUATCHED"`,
		},
		{
			name: "valid full config",
			cfg: Config{
				Version:        1,
				DefaultInstall: "steam",
				Installs: map[string]Install{
					"steam": {
						Path:             validPath,
						DefaultCharacter: "a1",
						Characters:       map[string]Character{"a1": {Name: "Squatched"}},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			err := Validate(&cfg)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "tilde expansion",
			in:   "~/games/ffxi/",
			want: filepath.Join(home, "games", "ffxi"),
		},
		{
			name:    "empty path",
			in:      "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizePath(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("trailing slash removed", func(t *testing.T) {
		dir := absPath(t, t.TempDir())
		got, err := NormalizePath(dir + string(filepath.Separator))
		if err != nil {
			t.Fatal(err)
		}
		if got != dir {
			t.Errorf("got %q, want %q", got, dir)
		}
	})
}

func TestResolveAlias(t *testing.T) {
	inst := &Install{
		Characters: map[string]Character{
			"a1b2c3d4": {Name: "Squatched"},
			"b2c3d4e5": {Name: "Alt"},
		},
	}

	tests := []struct {
		name    string
		inst    *Install
		query   string
		wantID  string
		wantErr string
	}{
		{name: "case insensitive", inst: inst, query: "squatched", wantID: "a1b2c3d4"},
		{name: "not found", inst: inst, query: "Nobody", wantErr: `no character found with name "Nobody"`},
		{name: "empty name", inst: inst, query: "  ", wantErr: "character name must not be empty"},
		{name: "nil install", inst: nil, query: "Squatched", wantErr: `no character found with name "Squatched"`},
		{name: "empty characters", inst: &Install{}, query: "Squatched", wantErr: `no character found with name "Squatched"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveAlias(tc.inst, tc.query)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.wantID {
				t.Errorf("id = %q, want %q", got, tc.wantID)
			}
		})
	}
}

func TestLookupName(t *testing.T) {
	inst := &Install{
		Characters: map[string]Character{
			"a1b2c3d4": {Name: "Squatched"},
		},
	}

	tests := []struct {
		inst *Install
		id   string
		want string
	}{
		{inst: inst, id: "a1b2c3d4", want: "Squatched"},
		{inst: inst, id: "missing", want: "missing"},
		{inst: nil, id: "a1b2c3d4", want: "a1b2c3d4"},
		{inst: &Install{Characters: map[string]Character{"a1": {Name: ""}}}, id: "a1", want: "a1"},
	}

	for _, tc := range tests {
		if got := LookupName(tc.inst, tc.id); got != tc.want {
			t.Errorf("LookupName(%v, %q) = %q, want %q", tc.inst, tc.id, got, tc.want)
		}
	}
}

func TestSuggestInstallName(t *testing.T) {
	cfg := Config{
		Version: 1,
		Installs: map[string]Install{
			"steam":   {Path: "/a"},
			"steam.2": {Path: "/b"},
		},
	}

	tests := []struct {
		path string
		want string
	}{
		{"/home/.steam/steamapps", "steam.3"},
		{"/home/Games/lutris/ffxi", "lutris"},
		{"/home/.wine/drive_c/ffxi", "wine"},
		{"/opt/playonline/ffxi", "playonline"},
		{"/mnt/games/ffxi", "default"},
	}

	for _, tc := range tests {
		if got := SuggestInstallName(&cfg, tc.path); got != tc.want {
			t.Errorf("SuggestInstallName(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestInstallNames(t *testing.T) {
	if got := InstallNames(&Config{}); got != nil {
		t.Errorf("empty config: got %v, want nil", got)
	}
	cfg := Config{
		Version: 1,
		Installs: map[string]Install{
			"lutris": {Path: "/a"},
			"steam":  {Path: "/b"},
		},
	}
	got := InstallNames(&cfg)
	want := []string{"lutris", "steam"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDefaultOffering(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{"absent", Config{}, true},
		{"explicit true", Config{Preferences: &Preferences{DefaultOffering: boolPtr(true)}}, true},
		{"explicit false", Config{Preferences: &Preferences{DefaultOffering: boolPtr(false)}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := DefaultOffering(&tc.cfg); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	valid := []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"YES", true},
		{"y", true},
		{"n", false},
		{"False", false},
		{"  no  ", false},
	}
	for _, tc := range valid {
		got, err := ParseBool(tc.in)
		if err != nil {
			t.Fatalf("ParseBool(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseBool(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}

	invalid := []string{"", "maybe", "2", "truthy"}
	for _, in := range invalid {
		if _, err := ParseBool(in); err == nil {
			t.Errorf("ParseBool(%q) expected error", in)
		}
	}
}

func TestFindInstallByPath(t *testing.T) {
	dir := t.TempDir()
	ffxi := absPath(t, dir, "ffxi")
	cfg := Config{
		Version: 1,
		Installs: map[string]Install{
			"test": {Path: ffxi},
		},
	}

	name, inst, err := FindInstallByPath(&cfg, ffxi)
	if err != nil {
		t.Fatal(err)
	}
	if name != "test" || inst == nil {
		t.Fatalf("got name=%q inst=%v", name, inst)
	}

	name, inst, err = FindInstallByPath(&cfg, absPath(t, dir, "other"))
	if err != nil {
		t.Fatal(err)
	}
	if name != "" || inst != nil {
		t.Errorf("unexpected match: name=%q inst=%v", name, inst)
	}
}

func TestPath_RespectsEnvOverride(t *testing.T) {
	t.Setenv("MACROMOG_CONFIG", "/custom/config.yml")
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/custom/config.yml" {
		t.Errorf("got %q, want /custom/config.yml", got)
	}
}

func boolPtr(b bool) *bool { return &b }