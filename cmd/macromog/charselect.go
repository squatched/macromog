package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/squatched/macromog/internal/aliases"
	"github.com/squatched/macromog/internal/lister"
)

// resolveCharDirs returns character USER directory paths to operate on.
//
// charDir and charName are mutually exclusive; passing both is an error.
// charDir, if non-empty, is validated as an existing directory and returned
// directly, bypassing discovery and prompting.
// charName, if non-empty, is resolved via USER/characters.yml to a hex ID.
// all returns every discovered character without prompting; it is incompatible
// with both charDir and charName.
// Otherwise the USER directory is scanned; if more than one character is found
// and stdin is a terminal, the user is prompted to pick one or more.
func resolveCharDirs(charDir, charName, ffxiPath string, all bool) ([]string, error) {
	if charDir != "" && charName != "" {
		return nil, fmt.Errorf("--char-dir and --char-name are mutually exclusive")
	}
	if charDir != "" && all {
		return nil, fmt.Errorf("--char-dir and --all are mutually exclusive")
	}
	if charName != "" && all {
		return nil, fmt.Errorf("--char-name and --all are mutually exclusive")
	}

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

	if charName != "" {
		doc, err := aliases.Load(userDir)
		if err != nil && !aliases.IsFutureVersion(err) {
			return nil, fmt.Errorf("loading aliases: %w", err)
		}
		hexID, err := aliases.Resolve(doc, charName)
		if err != nil {
			return nil, err
		}
		charDirPath := filepath.Join(userDir, hexID)
		if st, err := os.Stat(charDirPath); err != nil || !st.IsDir() {
			return nil, fmt.Errorf("character directory not found for %q: %s", charName, charDirPath)
		}
		return []string{charDirPath}, nil
	}

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", userDir, err)
	}
	if len(chars) == 0 {
		return nil, fmt.Errorf("no character directories found in %s; use --char-dir to specify one", userDir)
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
	return promptCharSelect(chars, userDir)
}

// resolveCharDir is a single-character convenience wrapper around
// resolveCharDirs for commands that always operate on exactly one character.
func resolveCharDir(charDir, charName, ffxiPath string) (string, error) {
	dirs, err := resolveCharDirs(charDir, charName, ffxiPath, false)
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

func promptCharSelect(chars []lister.CharacterInfo, userDir string) ([]string, error) {
	aliasDoc, _ := aliases.Load(userDir)

	if !stdinIsTerminal() {
		var sb strings.Builder
		fmt.Fprint(&sb, "multiple characters found; use --char-dir or --all to specify:")
		for _, c := range chars {
			name := aliases.LookupName(aliasDoc, c.ID)
			if name != c.ID {
				fmt.Fprintf(&sb, "\n  %s (%s)", name, c.ID)
			} else {
				fmt.Fprintf(&sb, "\n  %s", c.ID)
			}
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

	fmt.Fprintln(os.Stderr, "Multiple characters found. Select characters:")
	for i, c := range chars {
		suffix := "books"
		if c.BookCount == 1 {
			suffix = "book"
		}
		name := aliases.LookupName(aliasDoc, c.ID)
		if name != c.ID {
			fmt.Fprintf(os.Stderr, "  [%d] %s (%s) (%d %s)\n", i+1, name, c.ID, c.BookCount, suffix)
		} else {
			fmt.Fprintf(os.Stderr, "  [%d] %s (%d %s)\n", i+1, c.ID, c.BookCount, suffix)
		}
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
	return nil, fmt.Errorf("invalid selection; use --char-dir or --all to specify")
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
	return term.IsTerminal(int(os.Stdin.Fd()))
}
