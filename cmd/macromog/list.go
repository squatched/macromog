package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

const listUsage = `Usage: macromog list [flags]

List detected FFXI characters and their macro books.

Without --char-dir or --char-name, scans the FFXI USER directory and lists
every detected character with a populated-book count. Register character
aliases with 'macromog config set-alias' to display friendly names.

Flags:
  --ffxi-path <path>    FFXI install root (auto-detected if omitted)
  --install <name>      named FFXI install from config
  --char-dir <path>     character USER directory; lists books for that character
  --char-name <name>    friendly character name from config; lists books for that character

Examples:
  macromog list
  macromog list --install steam
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
	installName := fs.String("install", "", "named FFXI install from config")
	charDir := fs.String("char-dir", "", "character USER directory")
	charName := fs.String("char-name", "", "friendly character name from config")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *charDir != "" || *charName != "" {
		dir, err := resolveCharDir(charSelectOpts{
			charDir: *charDir, charName: *charName, ffxiPath: *ffxiPath, installName: *installName,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
			return 1
		}
		return runListChar(dir, p)
	}
	return runListAll(*ffxiPath, *installName, p)
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
	displayName := lookupCharName(filepath.Dir(charDirAbs), charID)

	p.Text(func(tw *TextWriter) {
		if displayName != charID {
			fmt.Fprintf(tw, "Character: %s %s\n\n", tw.Bold(displayName), tw.Muted("("+charID+")"))
		} else {
			fmt.Fprintf(tw, "Character: %s\n\n", tw.Highlight(charID))
		}
		if len(books) == 0 {
			fmt.Fprintln(tw, tw.Muted("  (no macros found)"))
			return
		}
		for _, b := range books {
			name := b.Name
			var coloredName string
			if name == "" {
				coloredName = tw.PadRight(tw.Muted("(unnamed)"), 16)
			} else {
				coloredName = tw.PadRight(tw.Cyan(name), 16)
			}
			sets := "sets"
			if b.SetCount == 1 {
				sets = "set"
			}
			fmt.Fprintf(tw, "  Book %s  %s  %d %s\n", tw.Muted(fmt.Sprintf("%2d", b.Index)), coloredName, b.SetCount, sets)
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

func runListAll(ffxiPath, installName string, p *Printer) int {
	userDir, inst, err := resolveUserDirForList(ffxiPath, installName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
		return 1
	}

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog list: %v\n", err)
		return 1
	}

	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "FFXI USER: %s\n\n", tw.Highlight(userDir))
		if len(chars) == 0 {
			fmt.Fprintln(tw, tw.Muted("  (no character directories found)"))
			return
		}
		for _, c := range chars {
			displayName := config.LookupName(inst, c.ID)
			var label string
			if displayName != c.ID {
				label = tw.PadRight(
					fmt.Sprintf("%s %s", tw.Bold(displayName), tw.Muted("("+c.ID+")")),
					28,
				)
			} else {
				label = tw.PadRight(tw.Highlight(c.ID), 28)
			}
			switch c.BookCount {
			case 0:
				fmt.Fprintf(tw, "  %s  %s\n", label, tw.Muted("(no macros)"))
			case 1:
				fmt.Fprintf(tw, "  %s  1 book\n", label)
			default:
				fmt.Fprintf(tw, "  %s  %d books\n", label, c.BookCount)
			}
		}
	})

	entries := make([]listCharEntry, len(chars))
	for i, c := range chars {
		entry := listCharEntry{ID: c.ID, BookCount: c.BookCount}
		if name := config.LookupName(inst, c.ID); name != c.ID {
			entry.Name = name
		}
		entries[i] = entry
	}
	p.JSON(listAllResult{UserDir: userDir, Characters: entries})

	return 0
}
