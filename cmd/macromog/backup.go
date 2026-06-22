package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/squatched/macromog/internal/backup"
)

const backupUsage = `Usage: macromog backup [flags] [<char-dir>]

Create a timestamped backup of all macro .dat files for a character.

Arguments:
  [<char-dir>]    character USER directory (required unless --char is given)

Flags:
  --char <path>   character USER directory (same as positional <char-dir>)

Examples:
  macromog backup /path/to/USER/a1b2c3d4
  macromog backup --char /path/to/USER/a1b2c3d4
`

func runBackup(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, backupUsage)
		return 0
	}

	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	charDir := fs.String("char", "", "character USER directory")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *charDir == "" && len(fs.Args()) > 0 {
		*charDir = fs.Args()[0]
	}

	if *charDir == "" {
		fmt.Fprint(os.Stderr, backupUsage)
		fmt.Fprintln(os.Stderr, "macromog backup: character directory required (<char-dir> or --char)")
		return 1
	}

	charDirAbs, err := filepath.Abs(*charDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
		return 1
	}
	if st, err := os.Stat(charDirAbs); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "macromog backup: character directory not found: %s\n", charDirAbs)
		return 1
	}

	backupDir, err := backup.Backup(charDirAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
		return 1
	}

	fmt.Printf("backed up to %s\n", backupDir)
	return 0
}
