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
  [<char-dir>]        character USER directory (auto-detected if omitted)

Flags:
  --ffxi-path <path>  FFXI install root (auto-detected if omitted)
  --char <path>       character USER directory; bypasses selection
  --all               back up all discovered characters without prompting
  --out <path>        directory to write the backup into (default: current directory)
  --in-place          write the backup into <char-dir>/backups/

Examples:
  macromog backup
  macromog backup /path/to/USER/a1b2c3d4
  macromog backup --all --in-place
  macromog backup --char /path/to/USER/a1b2c3d4 --out ~/macro-backups
`

func runBackup(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, backupUsage)
		return 0
	}

	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	charDir := fs.String("char", "", "character USER directory")
	all := fs.Bool("all", false, "back up all discovered characters")
	outDir := fs.String("out", "", "directory to write the backup into")
	inPlace := fs.Bool("in-place", false, "write backup into <char-dir>/backups/")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *charDir == "" && len(fs.Args()) > 0 {
		*charDir = fs.Args()[0]
	}

	if *outDir != "" && *inPlace {
		fmt.Fprintln(os.Stderr, "macromog backup: --out and --in-place are mutually exclusive")
		return 1
	}

	charDirs, err := resolveCharDirs(*charDir, *ffxiPath, *all)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
		return 1
	}

	var baseDestDir string
	if !*inPlace {
		if *outDir != "" {
			baseDestDir, err = filepath.Abs(*outDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
				return 1
			}
		} else {
			baseDestDir, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
				return 1
			}
		}
	}

	multi := len(charDirs) > 1
	failed := false
	for _, dir := range charDirs {
		destDir := baseDestDir
		if *inPlace {
			destDir = filepath.Join(dir, "backups")
		}
		backupDir, err := backup.Backup(dir, destDir)
		if err != nil {
			if multi {
				fmt.Fprintf(os.Stderr, "macromog backup: %s: %v\n", filepath.Base(dir), err)
			} else {
				fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
			}
			failed = true
			continue
		}
		if multi {
			fmt.Printf("[%s] backed up to %s\n", filepath.Base(dir), backupDir)
		} else {
			fmt.Printf("backed up to %s\n", backupDir)
		}
	}
	if failed {
		return 1
	}
	return 0
}
