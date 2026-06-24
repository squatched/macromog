package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RunningUnderWine reports whether the Windows binary is executing under Wine.
func RunningUnderWine() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	if os.Getenv("WINEPREFIX") != "" || os.Getenv("WINELOADER") != "" || os.Getenv("WINEARCH") != "" {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".config"))
	return err == nil
}

// CanonicalInstallPath normalizes an FFXI install root for storage in config.
// Under Wine the result is a host-native POSIX path so Linux and in-prefix
// Windows binaries can share one config file.
func CanonicalInstallPath(path string) (string, error) {
	if RunningUnderWine() {
		return canonicalForWine(path)
	}
	return NormalizePath(path)
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

func winePrefixDir() (string, error) {
	if p := strings.TrimSpace(os.Getenv("WINEPREFIX")); p != "" {
		return NormalizePath(expandHome(p))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wine"), nil
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
	// Map POSIX path to Wine's Z: drive (Unix root) for Windows API access.
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