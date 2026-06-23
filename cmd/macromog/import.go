package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/squatched/macromog/internal/importer"
)

const importUsage = `Usage: macromog import [flags] <file> [<char-dir>]

Import macros from YAML into FFXI .dat files.
A timestamped backup of current macros is created before writing (use
--no-backup to skip).

Arguments:
  <file>                YAML file to import (required)
  [<char-dir>]          character USER directory (auto-detected if omitted)

Flags:
  --ffxi-path <path>    FFXI install root (auto-detected if omitted)
  --char-dir <path>     character USER directory; bypasses selection
  --char-name <name>    character alias; bypasses selection
  --all                 import into all discovered characters without prompting
  --no-backup           skip the automatic backup before writing
  --dry-run             validate and show what would be written, without writing

Examples:
  macromog import mymacros.yml
  macromog import mymacros.yml /path/to/USER/a1b2c3d4
  macromog import --all mymacros.yml
  macromog import --char-dir /path/to/USER/a1b2c3d4 mymacros.yml
  macromog import --char-name Squatched mymacros.yml
  macromog import --dry-run mymacros.yml /path/to/USER/a1b2c3d4
`

type importEntry struct {
	Character  string          `json:"character"`
	YAMLFile   string          `json:"yaml_file"`
	Sets       int             `json:"sets"`
	BackupPath string          `json:"backup_path,omitempty"`
	DryRun     bool            `json:"dry_run"`
	DryRunSets []importSetInfo `json:"dry_run_sets,omitempty"`
	OK         bool            `json:"ok"`
	Error      string          `json:"error,omitempty"`
}

type importSetInfo struct {
	FileName string `json:"file"`
	Book     int    `json:"book"`
	Set      int    `json:"set"`
}

func runImport(args []string, p *Printer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, importUsage)
		return 0
	}

	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	charDir := fs.String("char-dir", "", "character USER directory")
	charName := fs.String("char-name", "", "character alias")
	all := fs.Bool("all", false, "import into all discovered characters")
	noBackup := fs.Bool("no-backup", false, "skip automatic backup")
	dryRun := fs.Bool("dry-run", false, "show what would be written without writing")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	remaining := fs.Args()
	var yamlPath string
	if len(remaining) > 0 {
		yamlPath = remaining[0]
		remaining = remaining[1:]
	}
	if *charDir == "" && *charName == "" && len(remaining) > 0 {
		*charDir = remaining[0]
	}

	if yamlPath == "" {
		fmt.Fprint(os.Stderr, importUsage)
		fmt.Fprintln(os.Stderr, "macromog import: YAML file required")
		return 1
	}

	charDirs, err := resolveCharDirs(*charDir, *charName, *ffxiPath, *all)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
		return 1
	}

	yamlAbs, err := filepath.Abs(yamlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
		return 1
	}

	multi := len(charDirs) > 1
	failed := false
	var results []importEntry

	for _, dir := range charDirs {
		charID := filepath.Base(dir)
		result, ierr := importer.Import(importer.Options{
			CharacterDir: dir,
			YAMLPath:     yamlAbs,
			Backup:       !*noBackup,
			DryRun:       *dryRun,
		})
		if ierr != nil {
			if !p.IsJSON() {
				if multi {
					fmt.Fprintf(os.Stderr, "macromog import: %s: %v\n", charID, ierr)
				} else {
					fmt.Fprintf(os.Stderr, "macromog import: %v\n", ierr)
				}
			}
			results = append(results, importEntry{
				Character: charID,
				YAMLFile:  filepath.Base(yamlPath),
				DryRun:    *dryRun,
				OK:        false,
				Error:     ierr.Error(),
			})
			failed = true
			continue
		}

		p.Text(func(tw *TextWriter) {
			if *dryRun {
				if multi {
					fmt.Fprintf(tw, "[%s] %s: would write %d macro set(s)\n", tw.Highlight(charID), tw.Warn("dry run"), len(result.Sets))
				} else {
					fmt.Fprintf(tw, "%s: would write %d macro set(s):\n", tw.Warn("dry run"), len(result.Sets))
					for _, s := range result.Sets {
						fmt.Fprintf(tw, "  %s %s\n", s.FileName, tw.Muted(fmt.Sprintf("(book %d, set %d)", s.Book, s.Set)))
					}
				}
			} else {
				if multi {
					if result.BackupDir != "" {
						fmt.Fprintf(tw, "[%s] backed up to %s\n", tw.Highlight(charID), tw.Muted(result.BackupDir))
					}
					fmt.Fprintf(tw, "[%s] imported %d macro set(s) from %s\n", tw.Highlight(charID), len(result.Sets), tw.Success(filepath.Base(yamlPath)))
				} else {
					if result.BackupDir != "" {
						fmt.Fprintf(tw, "backed up to %s\n", tw.Muted(result.BackupDir))
					}
					fmt.Fprintf(tw, "imported %d macro set(s) from %s\n", len(result.Sets), tw.Success(filepath.Base(yamlPath)))
				}
			}
		})

		entry := importEntry{
			Character: charID,
			YAMLFile:  filepath.Base(yamlPath),
			Sets:      len(result.Sets),
			DryRun:    *dryRun,
			OK:        true,
		}
		if result.BackupDir != "" {
			entry.BackupPath = result.BackupDir
		}
		if *dryRun {
			sets := make([]importSetInfo, len(result.Sets))
			for i, s := range result.Sets {
				sets[i] = importSetInfo{FileName: s.FileName, Book: s.Book, Set: s.Set}
			}
			entry.DryRunSets = sets
		}
		results = append(results, entry)
	}

	if p.IsJSON() {
		if multi {
			p.JSON(results)
		} else if len(results) == 1 {
			p.JSON(results[0])
		}
		if failed {
			for _, r := range results {
				if !r.OK {
					fmt.Fprintf(os.Stderr, "macromog import: %s: %s\n", r.Character, r.Error)
				}
			}
		}
	}

	if failed {
		return 1
	}
	return 0
}
