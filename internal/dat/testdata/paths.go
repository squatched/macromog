package testdata

import (
	"path/filepath"
	"runtime"
)

// CharDir returns the anonymized FFXI USER fixture directory used in tests.
func CharDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "char")
}
