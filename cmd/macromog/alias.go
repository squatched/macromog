package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/squatched/macromog/internal/aliases"
	"github.com/squatched/macromog/internal/lister"
)

const aliasUsage = `Usage: macromog alias [flags] <char-id> <name>
       macromog alias [flags] --remove <char-id>

Assign or remove a friendly name for an FFXI character folder.

The alias is stored in USER/characters.yml next to the character folders.
Other commands accept --char-name <name> to select a character by alias.

Arguments:
  <char-id>           hex character folder ID (e.g. c75a3f)
  <name>              friendly name to associate with the folder

Flags:
  --ffxi-path <path>  FFXI install root (auto-detected if omitted)
  --remove            remove the alias for <char-id> instead of setting it

Examples:
  macromog alias c75a3f Squatched
  macromog alias --remove c75a3f
  macromog alias --ffxi-path "/path/to/FINAL FANTASY XI" c75a3f Squatched
`

func runAlias(args []string, p *Printer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, aliasUsage)
		return 0
	}

	fs := flag.NewFlagSet("alias", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	remove := fs.Bool("remove", false, "remove alias for <char-id>")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	userDir, err := resolveUserDir(*ffxiPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}

	remaining := fs.Args()

	if *remove {
		if len(remaining) != 1 {
			fmt.Fprintln(os.Stderr, "macromog alias: --remove requires exactly one argument: <char-id>")
			return 1
		}
		return runAliasRemove(userDir, remaining[0], p)
	}

	if len(remaining) != 2 {
		fmt.Fprint(os.Stderr, aliasUsage)
		fmt.Fprintln(os.Stderr, "macromog alias: expected <char-id> and <name>")
		return 1
	}
	return runAliasSet(userDir, remaining[0], remaining[1], p)
}

type aliasSetResult struct {
	CharID string `json:"char_id"`
	Name   string `json:"name"`
}

type aliasRemoveResult struct {
	CharID  string `json:"char_id"`
	Removed bool   `json:"removed"`
}

func runAliasSet(userDir, charID, name string, p *Printer) int {
	if strings.TrimSpace(name) == "" {
		fmt.Fprintln(os.Stderr, "macromog alias: name must not be empty")
		return 1
	}
	charDir := filepath.Join(userDir, charID)
	if !lister.IsCharacterDir(charDir) {
		fmt.Fprintf(os.Stderr, "macromog alias: %q is not a valid character directory in %s\n", charID, userDir)
		return 1
	}

	doc, err := aliases.Load(userDir)
	if aliases.IsFutureVersion(err) {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}

	if doc.Chars == nil {
		doc.Chars = make(map[string]aliases.Entry)
	}
	doc.Chars[charID] = aliases.Entry{Name: name}

	if err := aliases.Save(userDir, doc); err != nil {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}

	p.Text(func(w io.Writer) {
		fmt.Fprintf(w, "alias set: %s → %s\n", charID, name)
	})
	p.JSON(aliasSetResult{CharID: charID, Name: name})
	return 0
}

func runAliasRemove(userDir, charID string, p *Printer) int {
	doc, err := aliases.Load(userDir)
	if aliases.IsFutureVersion(err) {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}

	if _, ok := doc.Chars[charID]; !ok {
		fmt.Fprintf(os.Stderr, "macromog alias: no alias set for %q\n", charID)
		return 1
	}

	delete(doc.Chars, charID)

	if err := aliases.Save(userDir, doc); err != nil {
		fmt.Fprintf(os.Stderr, "macromog alias: %v\n", err)
		return 1
	}

	p.Text(func(w io.Writer) {
		fmt.Fprintf(w, "alias removed: %s\n", charID)
	})
	p.JSON(aliasRemoveResult{CharID: charID, Removed: true})
	return 0
}
