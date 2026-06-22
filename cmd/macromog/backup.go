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
The backup directory is named <char-id>_YYYYMMDD_HHMMSS.

Arguments:
  [<char-dir>]    character USER directory (required unless --char is given)

Flags:
  --char <path>   character USER directory (same as positional <char-dir>)
  --out <path>    directory to write the backup into (default: current directory)
  --in-place      write the backup into <char-dir>/backups/ (same as import's auto-backup)

Examples:
  macromog backup /path/to/USER/a1b2c3d4
  macromog backup --char /path/to/USER/a1b2c3d4 --out ~/macro-backups
  macromog backup --char /path/to/USER/a1b2c3d4 --in-place
`

func runBackup(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, backupUsage)
		return 0
	}

	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	charDir := fs.String("char", "", "character USER directory")
	outDir := fs.String("out", "", "directory to write the backup into")
	inPlace := fs.Bool("in-place", false, "write backup into <char-dir>/backups/")

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
	if *outDir != "" && *inPlace {
		fmt.Fprintln(os.Stderr, "macromog backup: --out and --in-place are mutually exclusive")
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

	var destDir string
	switch {
	case *inPlace:
		destDir = filepath.Join(charDirAbs, "backups")
	case *outDir != "":
		destDir, err = filepath.Abs(*outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
			return 1
		}
	default:
		destDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
			return 1
		}
	}

	backupDir, err := backup.Backup(charDirAbs, destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
		return 1
	}

	fmt.Printf("backed up to %s\n", backupDir)
	return 0
}
