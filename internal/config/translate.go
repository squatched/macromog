package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

func (h *HostFS) canonicalInstall(path string) (string, error) {
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
		return h.canonicalDrivePath(path)
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

func (h *HostFS) canonicalDrivePath(path string) (string, error) {
	drive := strings.ToLower(string(path[0]))
	rest := strings.TrimLeft(path[2:], `\/`)
	rest = strings.ReplaceAll(rest, `\`, `/`)
	prefix, err := h.winePrefix()
	if err != nil {
		return "", err
	}
	driveDir := "drive_" + drive
	prefix = normalizeHostPath(prefix)
	joined := hostpath(prefix, driveDir, rest)
	return toStoredPath(joined)
}

func (h *HostFS) winePrefix() (string, error) {
	if h.WinePrefix != "" {
		return h.WinePrefix, nil
	}
	prefix, err := h.discoverWinePrefix()
	if err != nil {
		return "", err
	}
	h.WinePrefix = prefix
	return prefix, nil
}

func resolveForWine(stored string) (string, error) {
	stored = strings.TrimSpace(stored)
	slash := filepath.ToSlash(strings.ReplaceAll(stored, `\`, `/`))
	if !strings.HasPrefix(slash, "/") {
		return NormalizePath(stored)
	}
	rest := strings.TrimPrefix(slash, "/")
	return "Z:\\" + strings.ReplaceAll(rest, "/", "\\"), nil
}

func isZDrivePath(path string) bool {
	return len(path) >= 2 && (path[0] == 'Z' || path[0] == 'z') && path[1] == ':'
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
	slash := normalizeHostPath(expanded)
	if strings.HasPrefix(slash, "/home/") {
		return strings.TrimRight(slash, "/"), nil
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(filepath.ToSlash(abs), "/"), nil
}
