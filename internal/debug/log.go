package debug

import (
	"fmt"
	"os"
)

// Enabled reports whether verbose stderr diagnostics are active.
func Enabled() bool {
	return os.Getenv("MACROMOG_DEBUG") != ""
}

// Logf writes a formatted line to stderr when MACROMOG_DEBUG is set.
func Logf(format string, args ...any) {
	if !Enabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "macromog debug: "+format+"\n", args...)
}