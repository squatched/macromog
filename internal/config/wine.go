package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/squatched/macromog/internal/debug"
)

const (
	ffxiUserSuffix = "drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	zHomeRoot      = `Z:\home`
)

// RunningUnderWine reports whether the Windows binary is executing under Wine.
func RunningUnderWine() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	if v := os.Getenv("WINEPREFIX"); v != "" {
		debug.Logf("RunningUnderWine: WINEPREFIX=%q", v)
		return true
	}
	if os.Getenv("WINELOADER") != "" || os.Getenv("WINEARCH") != "" {
		return true
	}
	if _, ok := LinuxHomeForSharedConfig(); ok {
		return true
	}
	if st, err := os.Stat(zHomeRoot); err == nil && st.IsDir() {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".config"))
	return err == nil
}

// LinuxHomeForSharedConfig returns the host Linux home directory when the
// Windows binary runs under Wine with a mapped Unix home (typically Z:).
func LinuxHomeForSharedConfig() (string, bool) {
	if home, ok := linuxHomeFromEnvWINEPREFIX(); ok {
		return home, true
	}
	return discoverLinuxHomeViaZDrive()
}

// OpenPath returns a filesystem path suitable for os.Open on this runtime.
// Canonical POSIX paths under /home are mapped to Z: when needed under Wine.
func OpenPath(canonical string) (string, error) {
	if runtime.GOOS != "windows" || !strings.HasPrefix(filepath.ToSlash(canonical), "/home/") {
		debug.Logf("OpenPath: passthrough %q", canonical)
		return canonical, nil
	}
	if !RunningUnderWine() {
		debug.Logf("OpenPath: passthrough (not wine) %q", canonical)
		return canonical, nil
	}
	if _, err := os.Stat(canonical); err == nil {
		debug.Logf("OpenPath: posix stat ok %q", canonical)
		return canonical, nil
	}
	got, err := resolveForWine(canonical)
	debug.Logf("OpenPath: mapped %q -> %q err=%v", canonical, got, err)
	return got, err
}

// CanonicalInstallPath normalizes an FFXI install root for storage in config.
// Under Wine the result is a host-native POSIX path so Linux and in-prefix
// Windows binaries can share one config file.
func CanonicalInstallPath(path string) (string, error) {
	if RunningUnderWine() {
		got, err := canonicalForWine(path)
		debug.Logf("CanonicalInstallPath: in=%q out=%q err=%v", path, got, err)
		return got, err
	}
	got, err := NormalizePath(path)
	debug.Logf("CanonicalInstallPath: in=%q out=%q err=%v", path, got, err)
	return got, err
}

// ResolveInstallPath converts a stored install path for filesystem access in
// the current runtime.
func ResolveInstallPath(stored string) (string, error) {
	if stored == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if RunningUnderWine() {
		return resolveForWine(stored)
	}
	return NormalizePath(stored)
}

func linuxHomeFromEnvWINEPREFIX() (string, bool) {
	prefix := strings.TrimSpace(os.Getenv("WINEPREFIX"))
	if prefix == "" {
		return "", false
	}
	return linuxHomeFromPath(prefix)
}

func linuxHomeFromPath(path string) (string, bool) {
	slash := strings.ReplaceAll(strings.TrimSpace(path), `\`, `/`)
	if slash == "" {
		return "", false
	}
	if isZDrivePath(slash) {
		slash = slash[2:]
	}
	slash = strings.TrimLeft(slash, `/`)
	parts := strings.Split(slash, "/")
	if len(parts) >= 2 && parts[0] == "home" && parts[1] != "" {
		return "/" + filepath.Join(parts[0], parts[1]), true
	}
	return "", false
}

func discoverLinuxHomeViaZDrive() (string, bool) {
	entries, err := os.ReadDir(zHomeRoot)
	if err != nil {
		return "", false
	}

	var withMacromog, withConfig []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		posixHome := filepath.Join("/home", e.Name())
		if st, err := os.Stat(filepath.Join(posixHome, ".config", "macromog")); err == nil && st.IsDir() {
			withMacromog = append(withMacromog, posixHome)
		}
		if st, err := os.Stat(filepath.Join(posixHome, ".config")); err == nil && st.IsDir() {
			withConfig = append(withConfig, posixHome)
		}
	}
	if len(withMacromog) == 1 {
		return withMacromog[0], true
	}
	if len(withConfig) == 1 {
		return withConfig[0], true
	}
	return "", false
}

// WinePrefixDir returns the active Wine prefix for path canonicalization.
func WinePrefixDir() (string, error) {
	return winePrefixDir()
}

func winePrefixDir() (string, error) {
	if p := strings.TrimSpace(os.Getenv("WINEPREFIX")); p != "" {
		return normalizeWinePrefixPath(p)
	}
	if home, ok := LinuxHomeForSharedConfig(); ok {
		if prefix, ok := findWinePrefixUnderHome(home); ok {
			return prefix, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wine"), nil
}

func normalizeWinePrefixPath(p string) (string, error) {
	if isZDrivePath(p) {
		rest := strings.TrimLeft(p[2:], `\/`)
		p = "/" + strings.ReplaceAll(rest, `\`, `/`)
	}
	return NormalizePath(expandHome(p))
}

func findWinePrefixUnderHome(home string) (string, bool) {
	var candidates []string
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
			candidates = append(candidates, filepath.Join(games, name))
		}
	}
	candidates = append(candidates, filepath.Join(home, ".wine"))

	for _, p := range candidates {
		if st, err := os.Stat(filepath.Join(p, ffxiUserSuffix)); err == nil && st.IsDir() {
			return p, true
		}
	}
	for _, p := range candidates {
		if st, err := os.Stat(filepath.Join(p, "drive_c")); err == nil && st.IsDir() {
			return p, true
		}
	}
	return "", false
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

func canonicalForWine(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path must not be empty")
	}

	if strings.HasPrefix(path, "/") {
		return toStoredPath(path)
	}

	if isZDrivePath(path) {
		return canonicalZDrivePath(path)
	}

	if isDrivePath(path) {
		return canonicalDrivePath(path)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return toStoredPath(abs)
}

func canonicalZDrivePath(path string) (string, error) {
	rest := strings.TrimLeft(path[2:], `\/`)
	rest = strings.ReplaceAll(rest, `\`, `/`)
	return toStoredPath("/" + rest)
}

func canonicalDrivePath(path string) (string, error) {
	drive := strings.ToLower(string(path[0]))
	rest := strings.TrimLeft(path[2:], `\/`)
	rest = strings.ReplaceAll(rest, `\`, `/`)
	prefix, err := winePrefixDir()
	if err != nil {
		return "", err
	}
	driveDir := "drive_" + drive
	joined := filepath.Join(prefix, driveDir, rest)
	return toStoredPath(joined)
}

func isZDrivePath(path string) bool {
	return len(path) >= 2 && (path[0] == 'Z' || path[0] == 'z') && path[1] == ':'
}

func resolveForWine(stored string) (string, error) {
	stored = strings.TrimSpace(stored)
	if !strings.HasPrefix(stored, "/") {
		return NormalizePath(stored)
	}
	rest := strings.TrimPrefix(stored, "/")
	return "Z:\\" + strings.ReplaceAll(rest, "/", "\\"), nil
}

func isDrivePath(path string) bool {
	if len(path) < 3 {
		return false
	}
	if path[1] != ':' {
		return false
	}
	drive := path[0]
	if (drive < 'A' || drive > 'Z') && (drive < 'a' || drive > 'z') {
		return false
	}
	return path[2] == '\\' || path[2] == '/'
}

func toStoredPath(path string) (string, error) {
	expanded := expandHome(path)
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(filepath.ToSlash(abs), "/"), nil
}