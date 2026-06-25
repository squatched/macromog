package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/squatched/macromog/internal/debug"
)

const (
	ffxiUserSuffix = "drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER"
	zHomeRoot      = `Z:\home`
)

// HostFS captures the runtime filesystem view for one process: stored POSIX
// paths in YAML versus paths suitable for os.Open / ReadDir on this GOOS.
type HostFS struct {
	GOOS       string
	UnderWine  bool
	LinuxHome  string
	WinePrefix string
}

var activeHostFS atomic.Pointer[HostFS]

func init() {
	activeHostFS.Store(DetectHostFS())
}

// DetectHostFS probes the environment once and returns a HostFS snapshot.
func DetectHostFS() *HostFS {
	h := &HostFS{GOOS: runtime.GOOS}
	h.UnderWine = h.discoverUnderWine()
	if home, ok := h.discoverLinuxHome(); ok {
		h.LinuxHome = home
	}
	if prefix, err := h.discoverWinePrefix(); err == nil {
		h.WinePrefix = prefix
	}
	return h
}

// ActiveHostFS returns the process-wide HostFS (overridable in tests).
func ActiveHostFS() *HostFS {
	return activeHostFS.Load()
}

// SetHostFSForTest replaces ActiveHostFS for the duration of a test.
func SetHostFSForTest(h *HostFS) func() {
	prev := activeHostFS.Load()
	activeHostFS.Store(h)
	return func() { activeHostFS.Store(prev) }
}

// Stored canonicalizes a path for storage in config.yml.
func (h *HostFS) Stored(p string) (string, error) {
	if h.UnderWine {
		got, err := h.canonicalInstall(p)
		debug.Logf("HostFS.Stored: in=%q out=%q err=%v", p, got, err)
		return got, err
	}
	got, err := NormalizePath(p)
	debug.Logf("HostFS.Stored: in=%q out=%q err=%v", p, got, err)
	return got, err
}

// Access converts a stored/canonical path for filesystem operations.
func (h *HostFS) Access(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if h.UnderWine {
		got, err := resolveForWine(p)
		debug.Logf("HostFS.Access: mapped %q -> %q err=%v", p, got, err)
		return got, err
	}
	slash := normalizeHostPath(p)
	if h.GOOS == "windows" && strings.HasPrefix(slash, "/home/") {
		debug.Logf("HostFS.Access: passthrough (not wine) %q", slash)
		return slash, nil
	}
	got, err := NormalizePath(p)
	debug.Logf("HostFS.Access: normalized %q -> %q err=%v", p, got, err)
	return got, err
}

// ConfigPath returns the canonical config file location for this runtime.
func (h *HostFS) ConfigPath() (string, error) {
	if p := os.Getenv("MACROMOG_CONFIG"); p != "" {
		debug.Logf("HostFS.ConfigPath: MACROMOG_CONFIG override %q", p)
		return p, nil
	}
	if dir, ok := h.SharedConfigDir(); ok {
		p := normalizeHostPath(configFileInDir(dir))
		debug.Logf("HostFS.ConfigPath: shared XDG %q", p)
		return p, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(configDir, "macromog", "config.yml")
	debug.Logf("HostFS.ConfigPath: user config dir %q", path)
	return path, nil
}
