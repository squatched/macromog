package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/squatched/macromog/internal/debug"
)

// discoverUnderWine reports whether the Windows binary is executing under Wine.
func (h *HostFS) discoverUnderWine() bool {
	if h.GOOS != "windows" {
		return false
	}
	if v := os.Getenv("WINEPREFIX"); v != "" {
		debug.Logf("RunningUnderWine: WINEPREFIX=%q", v)
		return true
	}
	if os.Getenv("WINELOADER") != "" || os.Getenv("WINEARCH") != "" {
		return true
	}
	if home, ok := h.discoverLinuxHome(); ok && home != "" {
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

// discoverLinuxHome returns the host Linux home when Wine maps a Unix home (Z:).
func (h *HostFS) discoverLinuxHome() (string, bool) {
	if home, ok := linuxHomeFromEnvWINEPREFIX(); ok {
		return home, true
	}
	return discoverLinuxHomeViaZDrive()
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
		return hostpath("/home", parts[1]), true
	}
	return "", false
}

func discoverLinuxHomeViaZDrive() (string, bool) {
	entries, err := os.ReadDir(zHomeRoot)
	if err != nil {
		return "", false
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return pickLinuxHomeFromZUsers(names, hostHomeStat)
}

type hostHomeStatFn func(posixHome string, elem ...string) (bool, error)

func hostHomeStat(posixHome string, elem ...string) (bool, error) {
	p := hostpath(append([]string{posixHome}, elem...)...)
	st, err := os.Stat(p)
	return err == nil && st.IsDir(), err
}

func pickLinuxHomeFromZUsers(names []string, stat hostHomeStatFn) (string, bool) {
	var withMacromog, withConfig []string
	for _, name := range names {
		posixHome := hostpath("/home", name)
		if ok, _ := stat(posixHome, ".config", "macromog"); ok {
			withMacromog = append(withMacromog, posixHome)
		}
		if ok, _ := stat(posixHome, ".config"); ok {
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

// discoverWinePrefix returns the active Wine prefix for path canonicalization.
// Lutris layout: ~/Games/<slug>/pfx with FFXI USER under drive_c, else drive_c alone.
func (h *HostFS) discoverWinePrefix() (string, error) {
	if p := strings.TrimSpace(os.Getenv("WINEPREFIX")); p != "" {
		return normalizeWinePrefixPath(p)
	}
	if h.LinuxHome != "" {
		if prefix, ok := h.findWinePrefixUnderHome(h.LinuxHome); ok {
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
	p = strings.TrimSpace(p)
	if isZDrivePath(p) {
		rest := strings.TrimLeft(p[2:], `\/`)
		p = "/" + strings.ReplaceAll(rest, `\`, `/`)
	}
	slash := normalizeHostPath(p)
	if strings.HasPrefix(slash, "/home/") {
		return strings.TrimRight(slash, "/"), nil
	}
	return NormalizePath(expandHome(p))
}

func (h *HostFS) findWinePrefixUnderHome(home string) (string, bool) {
	home = normalizeHostPath(home)
	var candidates []string
	games := hostpath(home, "Games")
	gamesAccess, err := h.Access(games)
	if err != nil {
		return "", false
	}
	if entries, err := os.ReadDir(gamesAccess); err == nil {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, name := range names {
			candidates = append(candidates, hostpath(games, name))
		}
	}
	candidates = append(candidates, hostpath(home, ".wine"))

	for _, p := range candidates {
		access, err := h.Access(hostpath(p, ffxiUserSuffix))
		if err != nil {
			continue
		}
		if st, err := os.Stat(access); err == nil && st.IsDir() {
			return p, true
		}
	}
	for _, p := range candidates {
		access, err := h.Access(hostpath(p, "drive_c"))
		if err != nil {
			continue
		}
		if st, err := os.Stat(access); err == nil && st.IsDir() {
			return p, true
		}
	}
	return "", false
}

// SharedConfigDir returns the host XDG config directory when Linux and Wine
// should share one config file.
func (h *HostFS) SharedConfigDir() (string, bool) {
	if h.GOOS != "windows" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		return filepath.Join(home, ".config", "macromog"), true
	}
	if h.LinuxHome != "" {
		return hostpath(h.LinuxHome, ".config", "macromog"), true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	if st, err := os.Stat(filepath.Join(home, ".config")); err == nil && st.IsDir() {
		return filepath.Join(home, ".config", "macromog"), true
	}
	return "", false
}
