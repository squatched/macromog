package main

import (
	"flag"
	"fmt"
	"io"
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

func runList(args []string, p *Printer) int {
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
		return runListChar(*charDir, p)
	}
	return runListAll(*ffxiPath, p)
}

type listCharResult struct {
	Character string          `json:"character"`
	Books     []listBookEntry `json:"books"`
}

type listBookEntry struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	SetCount int    `json:"set_count"`
}

type listAllResult struct {
	UserDir    string          `json:"user_dir"`
	Characters []listCharEntry `json:"characters"`
}

type listCharEntry struct {
	ID        string `json:"id"`
	BookCount int    `json:"book_count"`
}

func runListChar(charDir string, p *Printer) int {
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

	charID := filepath.Base(charDirAbs)

	p.Text(func(w io.Writer) {
		fmt.Fprintf(w, "Character: %s\n\n", charID)
		if len(books) == 0 {
			fmt.Fprintln(w, "  (no macros found)")
			return
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
			fmt.Fprintf(w, "  Book %2d  %-16s  %d %s\n", b.Index, name, b.SetCount, sets)
		}
	})

	entries := make([]listBookEntry, len(books))
	for i, b := range books {
		entries[i] = listBookEntry{Index: b.Index, Name: b.Name, SetCount: b.SetCount}
	}
	p.JSON(listCharResult{Character: charID, Books: entries})

	return 0
}

func runListAll(ffxiPath string, p *Printer) int {
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

	p.Text(func(w io.Writer) {
		fmt.Fprintf(w, "FFXI USER: %s\n\n", userDir)
		if len(chars) == 0 {
			fmt.Fprintln(w, "  (no character directories found)")
			return
		}
		for _, c := range chars {
			switch c.BookCount {
			case 0:
				fmt.Fprintf(w, "  %-12s  (no macros)\n", c.ID)
			case 1:
				fmt.Fprintf(w, "  %-12s  1 book\n", c.ID)
			default:
				fmt.Fprintf(w, "  %-12s  %d books\n", c.ID, c.BookCount)
			}
		}
	})

	entries := make([]listCharEntry, len(chars))
	for i, c := range chars {
		entries[i] = listCharEntry{ID: c.ID, BookCount: c.BookCount}
	}
	p.JSON(listAllResult{UserDir: userDir, Characters: entries})

	return 0
}
