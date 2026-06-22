package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/squatched/macromog/internal/lister"
)

// resolveCharDirs returns character USER directory paths to operate on.
// If charDir is non-empty, it is validated and returned as the only result,
// bypassing discovery and prompting (for scripted/plugin use).
// If all is true, every discovered character is returned without prompting.
// Otherwise the USER directory is scanned; if more than one character is found
// and stdin is a terminal, the user is prompted to pick one or more.
func resolveCharDirs(charDir, ffxiPath string, all bool) ([]string, error) {
	if charDir != "" {
		abs, err := filepath.Abs(charDir)
		if err != nil {
			return nil, err
		}
		if st, err := os.Stat(abs); err != nil || !st.IsDir() {
			return nil, fmt.Errorf("character directory not found: %s", abs)
		}
		return []string{abs}, nil
	}

	userDir, err := resolveUserDir(ffxiPath)
	if err != nil {
		return nil, err
	}

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", userDir, err)
	}
	if len(chars) == 0 {
		return nil, fmt.Errorf("no character directories found in %s; use --char to specify one", userDir)
	}
	if all {
		dirs := make([]string, len(chars))
		for i, c := range chars {
			dirs[i] = c.Dir
		}
		return dirs, nil
	}
	if len(chars) == 1 {
		fmt.Fprintf(os.Stderr, "using character: %s\n", chars[0].ID)
		return []string{chars[0].Dir}, nil
	}
	return promptCharSelect(chars)
}

// resolveCharDir is a single-character convenience wrapper around
// resolveCharDirs for commands that always operate on exactly one character.
func resolveCharDir(charDir, ffxiPath string) (string, error) {
	dirs, err := resolveCharDirs(charDir, ffxiPath, false)
	if err != nil {
		return "", err
	}
	return dirs[0], nil
}

func resolveUserDir(ffxiPath string) (string, error) {
	if ffxiPath != "" {
		userDir := lister.UserDirFromFFXIPath(ffxiPath)
		if st, err := os.Stat(userDir); err != nil || !st.IsDir() {
			return "", fmt.Errorf("USER directory not found under %s", ffxiPath)
		}
		return userDir, nil
	}
	return lister.DetectUserDir()
}

func promptCharSelect(chars []lister.CharacterInfo) ([]string, error) {
	if !stdinIsTerminal() {
		var sb strings.Builder
		fmt.Fprint(&sb, "multiple characters found; use --char or --all to specify:")
		for _, c := range chars {
			fmt.Fprintf(&sb, "\n  %s", c.ID)
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

	fmt.Fprintln(os.Stderr, "Multiple characters found. Select characters:")
	for i, c := range chars {
		suffix := "books"
		if c.BookCount == 1 {
			suffix = "book"
		}
		fmt.Fprintf(os.Stderr, "  [%d] %s (%d %s)\n", i+1, c.ID, c.BookCount, suffix)
	}
	fmt.Fprintf(os.Stderr, "Enter numbers (e.g. 1, 1,3, 1-%d, all): ", len(chars))

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		indices, err := parseSelection(scanner.Text(), len(chars))
		if err == nil && len(indices) > 0 {
			dirs := make([]string, len(indices))
			for i, idx := range indices {
				dirs[i] = chars[idx].Dir
			}
			return dirs, nil
		}
	}
	return nil, fmt.Errorf("invalid selection; use --char or --all to specify")
}

// parseSelection parses a user selection string into 0-based indices.
// Supports: single numbers ("2"), comma-separated ("1,3"), ranges ("1-3"),
// and "all" to select everything up to max. Duplicates are silently dropped.
func parseSelection(input string, max int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "all" {
		indices := make([]int, max)
		for i := range indices {
			indices[i] = i
		}
		return indices, nil
	}

	seen := make(map[int]bool)
	var result []int

	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if dashIdx := strings.Index(part, "-"); dashIdx > 0 {
			lo, err1 := strconv.Atoi(strings.TrimSpace(part[:dashIdx]))
			hi, err2 := strconv.Atoi(strings.TrimSpace(part[dashIdx+1:]))
			if err1 != nil || err2 != nil || lo < 1 || hi > max || lo > hi {
				return nil, fmt.Errorf("invalid range: %q", part)
			}
			for i := lo; i <= hi; i++ {
				if !seen[i] {
					seen[i] = true
					result = append(result, i-1)
				}
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil || n < 1 || n > max {
				return nil, fmt.Errorf("invalid selection: %q", part)
			}
			if !seen[n] {
				seen[n] = true
				result = append(result, n-1)
			}
		}
	}

	sort.Ints(result)
	return result, nil
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
