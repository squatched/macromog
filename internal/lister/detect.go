package lister

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
	const rel = `PlayOnline\SquareEnix\FINAL FANTASY XI\USER`
	return []string{
		filepath.Join(`C:\Program Files (x86)`, rel),
		filepath.Join(`C:\Program Files`, rel),
	}
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
