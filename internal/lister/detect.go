package lister

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/squatched/macromog/internal/config"
)

// ffxiSteamAppID is the Steam AppID for FINAL FANTASY XI Ultimate Collection
// Seekers Edition (https://store.steampowered.com/app/230330/).
const ffxiSteamAppID = "230330"

// DetectUserDir attempts to find the FFXI USER directory by searching common
// install locations for the current OS. Returns an error if nothing is found.
func DetectUserDir() (string, error) {
	for _, p := range userDirCandidates() {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("FFXI USER directory not found; use --ffxi-path to specify the install path")
}

// UserDirFromFFXIPath returns the USER subdirectory under an FFXI install root.
func UserDirFromFFXIPath(ffxiPath string) string {
	return filepath.Join(ffxiPath, "USER")
}

func userDirCandidates() []string {
	if runtime.GOOS == "windows" {
		return windowsCandidates()
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return linuxCandidates(home)
}

func windowsCandidates() []string {
	const playOnlineRel = `PlayOnline\SquareEnix\FINAL FANTASY XI\USER`
	const steamGameRel = `steamapps\common\FINAL FANTASY XI Online\USER`
	candidates := []string{
		// Standard PlayOnline install locations (native and Steam via PlayOnline).
		filepath.Join(`C:\Program Files (x86)`, playOnlineRel),
		filepath.Join(`C:\Program Files`, playOnlineRel),
		// Steam native install locations (default Steam folder on 64-bit and 32-bit Windows).
		filepath.Join(`C:\Program Files (x86)\Steam`, steamGameRel),
		filepath.Join(`C:\Program Files\Steam`, steamGameRel),
	}
	if config.RunningUnderWine() {
		candidates = append(winePrefixCandidates(), candidates...)
	}
	return candidates
}

func winePrefixCandidates() []string {
	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	prefix := strings.TrimSpace(os.Getenv("WINEPREFIX"))
	if prefix != "" {
		if strings.HasPrefix(prefix, "/") {
			return []string{filepath.Join(prefix, "drive_c", driveRel)}
		}
	}
	if home, ok := config.LinuxHomeForSharedConfig(); ok {
		return prefixCandidatesUnderHome(home, driveRel)
	}
	return nil
}

func prefixCandidatesUnderHome(home, driveRel string) []string {
	var out []string
	games := filepath.Join(home, "Games")
	if entries, err := os.ReadDir(games); err == nil {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, name := range names {
			out = append(out, filepath.Join(games, name, "drive_c", driveRel))
		}
	}
	out = append(out, filepath.Join(home, ".wine", "drive_c", driveRel))
	return out
}

func linuxCandidates(home string) []string {
	const driveRel = "Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"

	var candidates []string

	// Default Wine prefix.
	candidates = append(candidates, filepath.Join(home, ".wine/drive_c", driveRel))

	// Lutris: prefixes stored under ~/Games/<game-name>/drive_c/
	lutrisDrive := filepath.Join("drive_c", driveRel)
	candidates = append(candidates, scanSubdirs(filepath.Join(home, "Games"), lutrisDrive)...)

	// Steam/Proton: prefix at compatdata/<AppID>/pfx/drive_c/
	pfxDrive := filepath.Join("pfx", "drive_c", driveRel)
	for _, steamRoot := range []string{
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".local", "share", "Steam"),
	} {
		compatData := filepath.Join(steamRoot, "steamapps", "compatdata")
		candidates = append(candidates, filepath.Join(compatData, ffxiSteamAppID, pfxDrive))
	}

	return candidates
}

// scanSubdirs returns path candidates of the form filepath.Join(base, name, suffix)
// for each subdirectory of base, sorted deterministically.
func scanSubdirs(base, suffix string) []string {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = filepath.Join(base, name, suffix)
	}
	return out
}
