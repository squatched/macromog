package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/squatched/macromog/internal/aliases"
	"github.com/squatched/macromog/internal/lister"
)

const listUsage = `Usage: macromog list [flags]

List detected FFXI characters and their macro books.

Without --char-dir or --char-name, scans the FFXI USER directory and lists
every detected character with a populated-book count. Set a friendly name
with 'macromog alias' to display it alongside the folder ID.

Flags:
  --ffxi-path <path>    FFXI install root (auto-detected if omitted)
  --char-dir <path>     character USER directory; lists books for that character
  --char-name <name>    character alias; lists books for that character

Examples:
  macromog list
  macromog list --ffxi-path "/path/to/FINAL FANTASY XI"
  macromog list --char-dir /path/to/USER/a1b2c3d4
  macromog list --char-name Squatched
`

func runList(args []string, p *Printer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, listUsage)
		return 0
	}

	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	charDir := fs.String("char-dir", "", "character USER directory")
	charName := fs.String("char-name", "", "character alias")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *charDir != "" || *charName != "" {
		dir, err := resolveCharDir(*charDir, *charName, *ffxiPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
			return 1
		}
		return runListChar(dir, *ffxiPath, p)
	}
	return runListAll(*ffxiPath, p)
}

type listCharResult struct {
	Character string          `json:"character"`
	Name      string          `json:"name,omitempty"`
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
	Name      string `json:"name,omitempty"`
	BookCount int    `json:"book_count"`
}

func runListChar(charDir, ffxiPath string, p *Printer) int {
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
	userDir := filepath.Dir(charDirAbs)
	aliasDoc, _ := aliases.Load(userDir)
	displayName := aliases.LookupName(aliasDoc, charID)

	p.Text(func(w io.Writer) {
		if displayName != charID {
			fmt.Fprintf(w, "Character: %s (%s)\n\n", displayName, charID)
		} else {
			fmt.Fprintf(w, "Character: %s\n\n", charID)
		}
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
	result := listCharResult{Character: charID, Books: entries}
	if displayName != charID {
		result.Name = displayName
	}
	p.JSON(result)

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

	aliasDoc, _ := aliases.Load(userDir)

	p.Text(func(w io.Writer) {
		fmt.Fprintf(w, "FFXI USER: %s\n\n", userDir)
		if len(chars) == 0 {
			fmt.Fprintln(w, "  (no character directories found)")
			return
		}
		for _, c := range chars {
			displayName := aliases.LookupName(aliasDoc, c.ID)
			label := c.ID
			if displayName != c.ID {
				label = fmt.Sprintf("%s (%s)", displayName, c.ID)
			}
			switch c.BookCount {
			case 0:
				fmt.Fprintf(w, "  %-28s  (no macros)\n", label)
			case 1:
				fmt.Fprintf(w, "  %-28s  1 book\n", label)
			default:
				fmt.Fprintf(w, "  %-28s  %d books\n", label, c.BookCount)
			}
		}
	})

	entries := make([]listCharEntry, len(chars))
	for i, c := range chars {
		entry := listCharEntry{ID: c.ID, BookCount: c.BookCount}
		if name := aliases.LookupName(aliasDoc, c.ID); name != c.ID {
			entry.Name = name
		}
		entries[i] = entry
	}
	p.JSON(listAllResult{UserDir: userDir, Characters: entries})

	return 0
}
