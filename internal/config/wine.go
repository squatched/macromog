package config

// RunningUnderWine reports whether the Windows binary is executing under Wine.
func RunningUnderWine() bool {
	return ActiveHostFS().UnderWine
}

// LinuxHomeForSharedConfig returns the host Linux home directory when the
// Windows binary runs under Wine with a mapped Unix home (typically Z:).
func LinuxHomeForSharedConfig() (string, bool) {
	h := DetectHostFS()
	return h.LinuxHome, h.LinuxHome != ""
}

// OpenPath returns a filesystem path suitable for os.Open on this runtime.
func OpenPath(canonical string) (string, error) {
	return ActiveHostFS().Access(canonical)
}

// CanonicalInstallPath normalizes an FFXI install root for storage in config.
func CanonicalInstallPath(path string) (string, error) {
	return ActiveHostFS().Stored(path)
}

// ResolveInstallPath converts a stored install path for filesystem access in
// the current runtime.
func ResolveInstallPath(stored string) (string, error) {
	return ActiveHostFS().Access(stored)
}

// WinePrefixDir returns the active Wine prefix for path canonicalization.
func WinePrefixDir() (string, error) {
	return ActiveHostFS().winePrefix()
}

// canonicalForWine re-detects the environment; used by unit tests after t.Setenv.
func canonicalForWine(path string) (string, error) {
	return DetectHostFS().canonicalInstall(path)
}

// findWinePrefixUnderHome re-detects the environment; used by unit tests.
func findWinePrefixUnderHome(home string) (string, bool) {
	return DetectHostFS().findWinePrefixUnderHome(home)
}