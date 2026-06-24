package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

type charSelectOpts struct {
	charDir     string
	charName    string
	ffxiPath    string
	installName string
	all         bool
}

// resolveCharDirs returns character USER directory paths to operate on.
func resolveCharDirs(opts charSelectOpts) ([]string, error) {
	if opts.charDir != "" && opts.charName != "" {
		return nil, fmt.Errorf("--char-dir and --char-name are mutually exclusive")
	}
	if opts.charDir != "" && opts.all {
		return nil, fmt.Errorf("--char-dir and --all are mutually exclusive")
	}
	if opts.charName != "" && opts.all {
		return nil, fmt.Errorf("--char-name and --all are mutually exclusive")
	}

	if opts.charDir != "" {
		abs, err := filepath.Abs(opts.charDir)
		if err != nil {
			return nil, err
		}
		if st, err := os.Stat(abs); err != nil || !st.IsDir() {
			return nil, fmt.Errorf("character directory not found: %s", abs)
		}
		return []string{abs}, nil
	}

	session, err := openConfig()
	if err != nil {
		return nil, err
	}
	instCtx, err := resolveInstall(session, installOpts{
		ffxiPath:    opts.ffxiPath,
		installName: opts.installName,
	})
	if err != nil {
		return nil, err
	}
	userDir := lister.UserDirFromFFXIPath(instCtx.ffxiPath)

	if opts.charName != "" {
		if instCtx.install == nil {
			return nil, fmt.Errorf("no character found with name %q", opts.charName)
		}
		hexID, err := config.ResolveAlias(instCtx.install, opts.charName)
		if err != nil {
			return nil, err
		}
		charDirPath := filepath.Join(userDir, hexID)
		if st, err := os.Stat(charDirPath); err != nil || !st.IsDir() {
			return nil, fmt.Errorf("character directory not found for %q: %s", opts.charName, charDirPath)
		}
		return []string{charDirPath}, nil
	}

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", userDir, err)
	}
	if opts.all {
		if len(chars) == 0 {
			return nil, fmt.Errorf("no character directories found in %s; use --char-dir to specify one", userDir)
		}
		dirs := make([]string, len(chars))
		for i, c := range chars {
			dirs[i] = c.Dir
		}
		return dirs, nil
	}

	if instCtx.install != nil && instCtx.install.DefaultCharacter != "" {
		id := instCtx.install.DefaultCharacter
		charDirPath := filepath.Join(userDir, id)
		if st, err := os.Stat(charDirPath); err == nil && st.IsDir() {
			display := config.LookupName(instCtx.install, id)
			ew := newErrWriter()
			if display != id {
				fmt.Fprintf(ew, "using character: %s %s\n", ew.Bold(display), ew.Muted("("+id+")"))
			} else {
				fmt.Fprintf(ew, "using character: %s\n", ew.Highlight(id))
			}
			return []string{charDirPath}, nil
		}
	}

	if instCtx.install != nil && len(instCtx.install.Characters) == 1 {
		for id := range instCtx.install.Characters {
			charDirPath := filepath.Join(userDir, id)
			if st, err := os.Stat(charDirPath); err == nil && st.IsDir() {
				display := config.LookupName(instCtx.install, id)
				ew := newErrWriter()
				if display != id {
					fmt.Fprintf(ew, "using character: %s %s\n", ew.Bold(display), ew.Muted("("+id+")"))
				} else {
					fmt.Fprintf(ew, "using character: %s\n", ew.Highlight(id))
				}
				return []string{charDirPath}, nil
			}
		}
	}

	if len(chars) == 0 {
		if instCtx.install != nil && len(instCtx.install.Characters) > 1 {
			return promptConfiguredCharSelect(instCtx.install, userDir, session)
		}
		return nil, fmt.Errorf("no character directories found in %s; use --char-dir to specify one", userDir)
	}
	if len(chars) == 1 {
		ew := newErrWriter()
		display := config.LookupName(instCtx.install, chars[0].ID)
		if display != chars[0].ID {
			fmt.Fprintf(ew, "using character: %s %s\n", ew.Bold(display), ew.Muted("("+chars[0].ID+")"))
		} else {
			fmt.Fprintf(ew, "using character: %s\n", ew.Highlight(chars[0].ID))
		}
		return []string{chars[0].Dir}, nil
	}
	return promptCharSelect(chars, instCtx.install, instCtx.installName, session)
}

// resolveCharDir is a single-character convenience wrapper around resolveCharDirs.
func resolveCharDir(opts charSelectOpts) (string, error) {
	opts.all = false
	dirs, err := resolveCharDirs(opts)
	if err != nil {
		return "", err
	}
	return dirs[0], nil
}

// resolveUserDirForList returns the USER directory using install resolution.
func resolveUserDirForList(ffxiPath, installName string) (string, *config.Install, error) {
	session, err := openConfig()
	if err != nil {
		return "", nil, err
	}
	instCtx, err := resolveInstall(session, installOpts{ffxiPath: ffxiPath, installName: installName})
	if err != nil {
		return "", nil, err
	}
	return lister.UserDirFromFFXIPath(instCtx.ffxiPath), instCtx.install, nil
}

func promptCharSelect(chars []lister.CharacterInfo, inst *config.Install, installName string, session *configSession) ([]string, error) {
	if !stdinIsTerminal() {
		var sb strings.Builder
		fmt.Fprint(&sb, "multiple characters found; use --char-dir or --all to specify:")
		for _, c := range chars {
			name := config.LookupName(inst, c.ID)
			if name != c.ID {
				fmt.Fprintf(&sb, "\n  %s (%s)", name, c.ID)
			} else {
				fmt.Fprintf(&sb, "\n  %s", c.ID)
			}
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

	ew := newErrWriter()
	fmt.Fprintln(ew, ew.Bold("Which character for this command?"))
	for i, c := range chars {
		suffix := "books"
		if c.BookCount == 1 {
			suffix = "book"
		}
		name := config.LookupName(inst, c.ID)
		index := ew.Muted(fmt.Sprintf("[%d]", i+1))
		bookCount := ew.Muted(fmt.Sprintf("(%d %s)", c.BookCount, suffix))
		if name != c.ID {
			fmt.Fprintf(ew, "  %s %s %s %s\n", index, ew.Bold(name), ew.Muted("("+c.ID+")"), bookCount)
		} else {
			fmt.Fprintf(ew, "  %s %s %s\n", index, ew.Highlight(c.ID), bookCount)
		}
	}
	fmt.Fprintf(ew, "Enter numbers (e.g. 1, 1,3, 1-%d, all): ", len(chars))

	if line, ok := readStdinLine(); ok {
		indices, err := parseSelection(line, len(chars))
		if err == nil && len(indices) > 0 {
			if config.DefaultOffering(&session.cfg) && inst != nil && inst.DefaultCharacter == "" && len(inst.Characters) > 0 {
				tip := fmt.Sprintf("\nTip: macromog config set-default-character %s", chars[indices[0]].ID)
				if installName != "" {
					tip += fmt.Sprintf(" --install %s", installName)
				}
				fmt.Fprintln(ew, tip+" skips this prompt.")
			}
			dirs := make([]string, len(indices))
			for i, idx := range indices {
				dirs[i] = chars[idx].Dir
			}
			return dirs, nil
		}
	}
	return nil, fmt.Errorf("invalid selection; use --char-dir or --all to specify")
}

func promptConfiguredCharSelect(inst *config.Install, userDir string, session *configSession) ([]string, error) {
	type entry struct {
		id   string
		name string
		dir  string
	}
	var entries []entry
	for id, ch := range inst.Characters {
		dir := filepath.Join(userDir, id)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			continue
		}
		entries = append(entries, entry{id: id, name: ch.Name, dir: dir})
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no character directories found in %s; use --char-dir to specify one", userDir)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	if !stdinIsTerminal() {
		return nil, fmt.Errorf("multiple characters configured; use --char-dir, --char-name, or --all to specify")
	}

	ew := newErrWriter()
	fmt.Fprintln(ew, ew.Bold("Which character for this command?"))
	for i, e := range entries {
		fmt.Fprintf(ew, "  %s %s %s\n", ew.Muted(fmt.Sprintf("[%d]", i+1)), ew.Bold(e.name), ew.Muted("("+e.id+")"))
	}
	fmt.Fprint(ew, "Enter number: ")

	if line, ok := readStdinLine(); ok {
		indices, err := parseSelection(line, len(entries))
		if err == nil && len(indices) == 1 {
			return []string{entries[indices[0]].dir}, nil
		}
	}
	return nil, fmt.Errorf("invalid selection; use --char-dir or --char-name to specify")
}

// parseSelection parses a user selection string into 0-based indices.
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

func lookupCharName(userDir, charID string) string {
	session, err := openConfig()
	if err != nil {
		return charID
	}
	ffxiPath := filepath.Dir(userDir)
	_, inst, _ := config.FindInstallByPath(&session.cfg, ffxiPath)
	return config.LookupName(inst, charID)
}
