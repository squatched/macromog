package config

import (
	"path"
	"strings"
)

// hostpath joins Linux host filesystem path elements with forward slashes on
// every GOOS. Use this instead of filepath.Join for /home/... paths; on
// Windows/Wine filepath.Join turns /home/user into \home\user.
func hostpath(elem ...string) string {
	return path.Clean(path.Join(elem...))
}

// normalizeHostPath converts backslashes to forward slashes and cleans the path.
func normalizeHostPath(p string) string {
	return path.Clean(strings.ReplaceAll(strings.TrimSpace(p), `\`, `/`))
}
