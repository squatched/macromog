package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/squatched/macromog/internal/export"
)

const exportUsage = `Usage: macromog export [flags] <char-dir> [output]

Export macros from FFXI .dat files to YAML.

Arguments:
  <char-dir>      character USER directory (required unless --char is given)
  [output]        output file; default: <character>_macros_<timestamp>.yml

Flags:
  --char <path>   character USER directory (same as positional <char-dir>)
  --output <file> output file (-o shorthand)
  --name <name>   character name for YAML metadata (default: directory name)

Examples:
  macromog export data/dats/c75afe
  macromog export data/dats/c75afe hendrimod.yml
  macromog export --char /path/to/USER/c75afe -o hendrimod.yml
`

func runExport(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, exportUsage)
		return 0
	}

	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	charDir := fs.String("char", "", "character USER directory")
	output := fs.String("output", "", "output YAML file")
	shortOut := fs.String("o", "", "output YAML file (shorthand)")
	charName := fs.String("name", "", "character name for YAML metadata")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	remaining := fs.Args()
	if *charDir == "" && len(remaining) > 0 {
		*charDir = remaining[0]
		remaining = remaining[1:]
	}

	if *output == "" {
		*output = *shortOut
	}
	if *output == "" && len(remaining) > 0 {
		*output = remaining[0]
	}

	if *charDir == "" {
		fmt.Fprint(os.Stderr, exportUsage)
		fmt.Fprintln(os.Stderr, "macromog export: character directory required (<char-dir> or --char)")
		return 1
	}

	charDirAbs, err := filepath.Abs(*charDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog export: %v\n", err)
		return 1
	}
	if st, err := os.Stat(charDirAbs); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "macromog export: character directory not found: %s\n", charDirAbs)
		return 1
	}

	name := *charName
	if name == "" {
		name = filepath.Base(charDirAbs)
	}

	outPath := *output
	if outPath == "" {
		stamp := time.Now().UTC().Format("20060102_150405")
		outPath = fmt.Sprintf("%s_macros_%s.yml", strings.ToLower(name), stamp)
	}

	if err := export.WriteFile(charDirAbs, outPath, name); err != nil {
		fmt.Fprintf(os.Stderr, "macromog export: %v\n", err)
		return 1
	}

	fmt.Printf("exported macros to %s\n", outPath)
	return 0
}
