package lister_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/lister"
)

func TestUserDirFromFFXIPath(t *testing.T) {
	got := lister.UserDirFromFFXIPath("/path/to/FFXI")
	want := filepath.Join("/path/to/FFXI", "USER")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDetectUserDir_NotFound(t *testing.T) {
	// Override HOME to a temp dir with no FFXI install so auto-detection fails.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := lister.DetectUserDir()
	if err == nil {
		t.Fatal("expected error when FFXI not installed")
	}
}

func TestDetectUserDir_LinuxGames(t *testing.T) {
	// Create a fake wine prefix under ~/Games/mygame/drive_c/...
	home := t.TempDir()
	t.Setenv("HOME", home)

	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	userDir := filepath.Join(home, "Games", "mygame", "drive_c", driveRel)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := lister.DetectUserDir()
	if err != nil {
		t.Fatalf("DetectUserDir() unexpected error: %v", err)
	}
	if got != userDir {
		t.Errorf("got %q, want %q", got, userDir)
	}
}

func TestDetectUserDir_WineDefault(t *testing.T) {
	// Create a fake wine prefix under ~/.wine/drive_c/...
	home := t.TempDir()
	t.Setenv("HOME", home)

	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	userDir := filepath.Join(home, ".wine", "drive_c", driveRel)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := lister.DetectUserDir()
	if err != nil {
		t.Fatalf("DetectUserDir() unexpected error: %v", err)
	}
	if got != userDir {
		t.Errorf("got %q, want %q", got, userDir)
	}
}

func TestDetectUserDir_SteamProton(t *testing.T) {
	// Create a fake Proton prefix under ~/.steam/steam/steamapps/compatdata/<AppID>/pfx/
	home := t.TempDir()
	t.Setenv("HOME", home)

	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	userDir := filepath.Join(home, ".steam", "steam", "steamapps", "compatdata",
		"230330", "pfx", "drive_c", driveRel)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := lister.DetectUserDir()
	if err != nil {
		t.Fatalf("DetectUserDir() unexpected error: %v", err)
	}
	if got != userDir {
		t.Errorf("got %q, want %q", got, userDir)
	}
}

func TestDetectUserDir_SteamLocalShare(t *testing.T) {
	// Create a fake Proton prefix under ~/.local/share/Steam/steamapps/compatdata/<AppID>/pfx/
	home := t.TempDir()
	t.Setenv("HOME", home)

	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	userDir := filepath.Join(home, ".local", "share", "Steam", "steamapps", "compatdata",
		"230330", "pfx", "drive_c", driveRel)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := lister.DetectUserDir()
	if err != nil {
		t.Fatalf("DetectUserDir() unexpected error: %v", err)
	}
	if got != userDir {
		t.Errorf("got %q, want %q", got, userDir)
	}
}
