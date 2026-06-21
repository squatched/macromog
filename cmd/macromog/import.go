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
  <file>          YAML file to import (required)
  [<char-dir>]    character USER directory (required unless --char is given)

Flags:
  --char <path>   character USER directory (same as positional <char-dir>)
  --no-backup     skip the automatic backup before writing
  --dry-run       validate and show what would be written, without writing

Examples:
  macromog import mymacros.yml /path/to/USER/a1b2c3d4
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
	charDir := fs.String("char", "", "character USER directory")
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
	if *charDir == "" {
		fmt.Fprint(os.Stderr, importUsage)
		fmt.Fprintln(os.Stderr, "macromog import: character directory required (<char-dir> or --char)")
		return 1
	}

	charDirAbs, err := filepath.Abs(*charDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
		return 1
	}
	if st, err := os.Stat(charDirAbs); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "macromog import: character directory not found: %s\n", charDirAbs)
		return 1
	}

	yamlAbs, err := filepath.Abs(yamlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
		return 1
	}

	result, err := importer.Import(importer.Options{
		CharacterDir: charDirAbs,
		YAMLPath:     yamlAbs,
		Backup:       !*noBackup,
		DryRun:       *dryRun,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
		return 1
	}

	if *dryRun {
		fmt.Printf("dry run: would write %d macro set(s):\n", len(result.Sets))
		for _, s := range result.Sets {
			fmt.Printf("  %s (book %d, set %d)\n", s.FileName, s.Book, s.Set)
		}
		return 0
	}

	if result.BackupDir != "" {
		fmt.Printf("backed up to %s\n", result.BackupDir)
	}
	fmt.Printf("imported %d macro set(s) from %s\n", len(result.Sets), filepath.Base(yamlPath))
	return 0
}
