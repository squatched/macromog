package main

import (
	"bufio"
	"io"
	"os"

	"golang.org/x/term"
)

// interactiveStdin holds injectable stdin behavior. Tests override isTTY and r
// to exercise interactive prompts without a real terminal (same split as
// detectColor vs TextWriter{color: ...} for output).
var interactiveStdin = struct {
	isTTY   func() bool
	r       io.Reader
	scanner *bufio.Scanner
}{
	isTTY: defaultStdinIsTTY,
	r:     os.Stdin,
}

// isCI reports whether the CI environment variable is set (non-empty).
// CI=1 is the convention used by GitHub Actions, Jenkins, and other automation
// environments to signal that the process is running non-interactively.
func isCI() bool {
	return os.Getenv("CI") != ""
}

func defaultStdinIsTTY() bool {
	if isCI() {
		return false
	}
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func stdinIsTerminal() bool {
	return interactiveStdin.isTTY()
}

func stdinReader() io.Reader {
	return interactiveStdin.r
}

func readStdinLine() (string, bool) {
	if interactiveStdin.scanner == nil {
		interactiveStdin.scanner = bufio.NewScanner(stdinReader())
	}
	if !interactiveStdin.scanner.Scan() {
		return "", false
	}
	return interactiveStdin.scanner.Text(), true
}

func restoreInteractiveStdin() {
	interactiveStdin.isTTY = defaultStdinIsTTY
	interactiveStdin.r = os.Stdin
	interactiveStdin.scanner = nil
}
