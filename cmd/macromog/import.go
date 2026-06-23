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
  <file>              YAML file to import (required)
  [<char-dir>]        character USER directory (auto-detected if omitted)

Flags:
  --ffxi-path <path>  FFXI install root (auto-detected if omitted)
  --char <path>       character USER directory; bypasses selection
  --all               import into all discovered characters without prompting
  --no-backup         skip the automatic backup before writing
  --dry-run           validate and show what would be written, without writing

Examples:
  macromog import mymacros.yml
  macromog import mymacros.yml /path/to/USER/a1b2c3d4
  macromog import --all mymacros.yml
  macromog import --char /path/to/USER/a1b2c3d4 mymacros.yml
  macromog import --dry-run mymacros.yml /path/to/USER/a1b2c3d4
`

func runImport(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, importUsage)
		return 0
	}

	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	charDir := fs.String("char", "", "character USER directory")
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
	if *charDir == "" && len(remaining) > 0 {
		*charDir = remaining[0]
	}

	if yamlPath == "" {
		fmt.Fprint(os.Stderr, importUsage)
		fmt.Fprintln(os.Stderr, "macromog import: YAML file required")
		return 1
	}

	charDirs, err := resolveCharDirs(*charDir, *ffxiPath, *all)
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
	for _, dir := range charDirs {
		result, err := importer.Import(importer.Options{
			CharacterDir: dir,
			YAMLPath:     yamlAbs,
			Backup:       !*noBackup,
			DryRun:       *dryRun,
		})
		if err != nil {
			if multi {
				fmt.Fprintf(os.Stderr, "macromog import: %s: %v\n", filepath.Base(dir), err)
			} else {
				fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
			}
			failed = true
			continue
		}

		if *dryRun {
			if multi {
				fmt.Printf("[%s] dry run: would write %d macro set(s)\n", filepath.Base(dir), len(result.Sets))
			} else {
				fmt.Printf("dry run: would write %d macro set(s):\n", len(result.Sets))
				for _, s := range result.Sets {
					fmt.Printf("  %s (book %d, set %d)\n", s.FileName, s.Book, s.Set)
				}
			}
			continue
		}

		if multi {
			if result.BackupDir != "" {
				fmt.Printf("[%s] backed up to %s\n", filepath.Base(dir), result.BackupDir)
			}
			fmt.Printf("[%s] imported %d macro set(s) from %s\n", filepath.Base(dir), len(result.Sets), filepath.Base(yamlPath))
		} else {
			if result.BackupDir != "" {
				fmt.Printf("backed up to %s\n", result.BackupDir)
			}
			fmt.Printf("imported %d macro set(s) from %s\n", len(result.Sets), filepath.Base(yamlPath))
		}
	}
	if failed {
		return 1
	}
	return 0
}
