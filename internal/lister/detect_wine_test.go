package lister

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

func TestWinePrefixCandidates_MapsPOSIXToZ(t *testing.T) {
	home := t.TempDir()
	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	prefix := filepath.Join(home, "Games", "ffxi")
	userDir := filepath.Join(prefix, "drive_c", driveRel)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WINEPREFIX", prefix)

	wine := &config.HostFS{GOOS: "windows", UnderWine: true, LinuxHome: home}
	restore := config.SetHostFSForTest(wine)
	defer restore()

	got := winePrefixCandidates()
	if len(got) != 1 {
		t.Fatalf("winePrefixCandidates() = %v, want one candidate", got)
	}
	if !strings.HasPrefix(got[0], `Z:\`) {
		t.Errorf("candidate = %q, want Z:\\ prefix", got[0])
	}
	if strings.Contains(got[0], "/") {
		t.Errorf("candidate must not contain POSIX slashes: %q", got[0])
	}
}
