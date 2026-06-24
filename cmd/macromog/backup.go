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
  [<char-dir>]          character USER directory (auto-detected if omitted)

Flags:
  --ffxi-path <path>    FFXI install root (auto-detected if omitted)
  --char-dir <path>     character USER directory; bypasses selection
  --char-name <name>    character alias; bypasses selection
  --all                 back up all discovered characters without prompting
  --out <path>          directory to write the backup into (default: current directory)
  --in-place            write the backup into <char-dir>/backups/

Examples:
  macromog backup
  macromog backup /path/to/USER/a1b2c3d4
  macromog backup --all --in-place
  macromog backup --char-dir /path/to/USER/a1b2c3d4 --out ~/macro-backups
  macromog backup --char-name Squatched --out ~/macro-backups
`

type backupEntry struct {
	Character string `json:"character"`
	Path      string `json:"path,omitempty"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

func (r backupEntry) ok() bool          { return r.OK }
func (r backupEntry) character() string { return r.Character }
func (r backupEntry) errMsg() string    { return r.Error }

func runBackup(args []string, p *Printer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, backupUsage)
		return 0
	}

	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	charDir := fs.String("char-dir", "", "character USER directory")
	charName := fs.String("char-name", "", "character alias")
	all := fs.Bool("all", false, "back up all discovered characters")
	outDir := fs.String("out", "", "directory to write the backup into")
	inPlace := fs.Bool("in-place", false, "write backup into <char-dir>/backups/")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *charDir == "" && *charName == "" && len(fs.Args()) > 0 {
		*charDir = fs.Args()[0]
	}

	if *outDir != "" && *inPlace {
		fmt.Fprintln(os.Stderr, "macromog backup: --out and --in-place are mutually exclusive")
		return 1
	}

	charDirs, err := resolveCharDirs(*charDir, *charName, *ffxiPath, *all)
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
	var results []backupEntry

	for _, dir := range charDirs {
		charID := filepath.Base(dir)
		destDir := baseDestDir
		if *inPlace {
			destDir = filepath.Join(dir, "backups")
		}
		backupDir, berr := backup.Backup(dir, destDir)
		if berr != nil {
			if !p.IsJSON() {
				ew := p.Err()
				if multi {
					fmt.Fprintf(ew, "macromog backup: %s: %v\n", ew.Highlight(charID), berr)
				} else {
					fmt.Fprintf(ew, "macromog backup: %v\n", berr)
				}
			}
			results = append(results, backupEntry{Character: charID, OK: false, Error: berr.Error()})
			failed = true
			continue
		}

		p.Text(func(tw *TextWriter) {
			if multi {
				fmt.Fprintf(tw, "[%s] backed up to %s\n", tw.Highlight(charID), tw.Success(backupDir))
			} else {
				fmt.Fprintf(tw, "backed up to %s\n", tw.Success(backupDir))
			}
		})
		results = append(results, backupEntry{Character: charID, Path: backupDir, OK: true})
	}

	emitJSONResults(p, results, multi, failed, "backup")

	if failed {
		return 1
	}
	return 0
}
