package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/squatched/macromog/internal/lister"
)

const listUsage = `Usage: macromog list [flags]

List detected FFXI characters and their macro books.

Without --char, scans the FFXI USER directory (auto-detected or via
--ffxi-path) and lists every detected character with a populated-book count.

With --char, lists the macro books for that specific character directory.

Flags:
  --ffxi-path <path>  FFXI install root (auto-detected if omitted)
  --char <path>       character USER directory; lists books for that character

Examples:
  macromog list
  macromog list --ffxi-path "/path/to/FINAL FANTASY XI"
  macromog list --char /path/to/USER/a1b2c3d4
`

func runList(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, listUsage)
		return 0
	}

	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	charDir := fs.String("char", "", "character USER directory")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *charDir != "" {
		return runListChar(*charDir)
	}
	return runListAll(*ffxiPath)
}

func runListChar(charDir string) int {
	charDirAbs, err := filepath.Abs(charDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
		return 1
	}
	if st, err := os.Stat(charDirAbs); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "macromog list: character directory not found: %s\n", charDirAbs)
		return 1
	}

	books, err := lister.BooksForCharacter(charDirAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
		return 1
	}

	fmt.Printf("Character: %s\n\n", filepath.Base(charDirAbs))
	if len(books) == 0 {
		fmt.Println("  (no macros found)")
		return 0
	}
	for _, b := range books {
		name := b.Name
		if name == "" {
			name = "(unnamed)"
		}
		sets := "sets"
		if b.SetCount == 1 {
			sets = "set"
		}
		fmt.Printf("  Book %2d  %-16s  %d %s\n", b.Index, name, b.SetCount, sets)
	}
	return 0
}

func runListAll(ffxiPath string) int {
	var userDir string
	if ffxiPath != "" {
		userDir = lister.UserDirFromFFXIPath(ffxiPath)
		if st, err := os.Stat(userDir); err != nil || !st.IsDir() {
			fmt.Fprintf(os.Stderr, "macromog list: USER directory not found under %s\n", ffxiPath)
			return 1
		}
	} else {
		detected, err := lister.DetectUserDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
			return 1
		}
		userDir = detected
	}

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
		return 1
	}

	fmt.Printf("FFXI USER: %s\n\n", userDir)
	if len(chars) == 0 {
		fmt.Println("  (no character directories found)")
		return 0
	}
	for _, c := range chars {
		switch c.BookCount {
		case 0:
			fmt.Printf("  %-12s  (no macros)\n", c.ID)
		case 1:
			fmt.Printf("  %-12s  1 book\n", c.ID)
		default:
			fmt.Printf("  %-12s  %d books\n", c.ID, c.BookCount)
		}
	}
	return 0
}
