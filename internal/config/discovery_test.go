package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPickLinuxHomeFromZUsers_PrefersMacromog(t *testing.T) {
	stat := func(home string, elem ...string) (bool, error) {
		if len(elem) == 2 && elem[0] == ".config" && elem[1] == "macromog" {
			return home == "/home/alpha", nil
		}
		if len(elem) == 1 && elem[0] == ".config" {
			return home == "/home/alpha" || home == "/home/beta", nil
		}
		return false, nil
	}
	got, ok := pickLinuxHomeFromZUsers([]string{"alpha", "beta"}, stat)
	if !ok || got != "/home/alpha" {
		t.Fatalf("got %q ok=%v, want /home/alpha true", got, ok)
	}
}

func TestPickLinuxHomeFromZUsers_SingleConfig(t *testing.T) {
	stat := func(home string, elem ...string) (bool, error) {
		if len(elem) == 3 && elem[2] == "macromog" {
			return false, nil
		}
		return len(elem) == 1 && elem[0] == ".config" && home == "/home/solo", nil
	}
	got, ok := pickLinuxHomeFromZUsers([]string{"solo", "other"}, stat)
	if !ok || got != "/home/solo" {
		t.Fatalf("got %q ok=%v, want /home/solo true", got, ok)
	}
}

func TestPickLinuxHomeFromZUsers_Ambiguous(t *testing.T) {
	stat := func(home string, elem ...string) (bool, error) {
		return len(elem) == 1 && elem[0] == ".config", nil
	}
	_, ok := pickLinuxHomeFromZUsers([]string{"a", "b"}, stat)
	if ok {
		t.Fatal("expected false for multiple .config homes")
	}
}

func TestNormalizeWinePrefixPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "posix under home",
			in:   "/home/squatched/Games/final-fantasy-xi-online/pfx",
			want: "/home/squatched/Games/final-fantasy-xi-online/pfx",
		},
		{
			name: "z drive prefix",
			in:   `Z:\home\squatched\Games\final-fantasy-xi-online\pfx`,
			want: "/home/squatched/Games/final-fantasy-xi-online/pfx",
		},
		{
			name: "trailing slash trimmed",
			in:   "/home/squatched/.wine/",
			want: "/home/squatched/.wine",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeWinePrefixPath(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFindWinePrefixUnderHome_WineAccessMapping(t *testing.T) {
	home := t.TempDir()
	prefix := filepath.Join(home, "Games", "lutris-game")
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
	games := hostpath(home, "Games")
	access, err := wine.Access(games)
	if err != nil {
		t.Fatal(err)
	}
	wantZ := `Z:\` + strings.ReplaceAll(strings.TrimPrefix(games, "/"), "/", `\`)
	if access != wantZ {
		t.Errorf("Access(games) = %q, want %q", access, wantZ)
	}

	// Full prefix discovery under UnderWine requires a live Z: tree (see validate-wine-smoke).
}
