package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
)

const exportUsage = `Usage: macromog export [flags] [<char-dir>] [output]

Export macros from FFXI .dat files to YAML.

Arguments:
  [<char-dir>]          character USER directory (auto-detected if omitted)
  [output]              output file; default: <character>_macros_<timestamp>.yml
                        not valid when multiple characters are selected

Flags:
  --ffxi-path <path>    FFXI install root (auto-detected if omitted)
  --install <name>      named FFXI install from config
  --char-dir <path>     character USER directory; bypasses selection
  --char-name <name>    friendly character name from config; bypasses selection
  --all                 export all discovered characters without prompting
  --output <file>       output YAML file (-o shorthand); requires one character
  --name <name>         character name for YAML metadata; requires one character
  --scope <selector>    scope selector (repeatable; e.g. B1S3, B1,5, B1S3A1)

Examples:
  macromog export
  macromog export --all
  macromog export /path/to/USER/a1b2c3d4
  macromog export /path/to/USER/a1b2c3d4 macros.yml
  macromog export --char-dir /path/to/USER/a1b2c3d4 -o macros.yml
  macromog export --char-name Squatched -o macros.yml
  macromog export --scope B1,5 -o books1and5.yml
`

type exportEntry struct {
	Character string `json:"character"`
	Path      string `json:"path,omitempty"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

func (r exportEntry) ok() bool          { return r.OK }
func (r exportEntry) character() string { return r.Character }
func (r exportEntry) errMsg() string    { return r.Error }

func runExport(args []string, p *Printer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, exportUsage)
		return 0
	}

	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	ffxiPath := fs.String("ffxi-path", "", "FFXI install root")
	installName := fs.String("install", "", "named FFXI install from config")
	charDir := fs.String("char-dir", "", "character USER directory")
	charName := fs.String("char-name", "", "character alias")
	all := fs.Bool("all", false, "export all discovered characters")
	output := fs.String("output", "", "output YAML file")
	shortOut := fs.String("o", "", "output YAML file (shorthand)")
	metaName := fs.String("name", "", "character name for YAML metadata")
	var scopeSel scopeFlags
	fs.Var(&scopeSel, "scope", "scope selector (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	sc, err := scope.ParseSelectors([]string(scopeSel))
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog export: invalid --scope: %v\n", err)
		return 1
	}

	remaining := fs.Args()
	if *charDir == "" && *charName == "" && len(remaining) > 0 {
		*charDir = remaining[0]
		remaining = remaining[1:]
	}

	if *output == "" {
		*output = *shortOut
	}
	if *output == "" && len(remaining) > 0 {
		*output = remaining[0]
	}

	charDirs, err := resolveCharDirs(charSelectOpts{
		charDir: *charDir, charName: *charName, ffxiPath: *ffxiPath,
		installName: *installName, all: *all,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog export: %v\n", err)
		return 1
	}

	if len(charDirs) > 1 && *output != "" {
		fmt.Fprintln(os.Stderr, "macromog export: --output requires exactly one character; use --char-dir or remove --all")
		return 1
	}
	if len(charDirs) > 1 && *metaName != "" {
		fmt.Fprintln(os.Stderr, "macromog export: --name requires exactly one character; use --char-dir or remove --all")
		return 1
	}

	stamp := time.Now().UTC().Format("20060102_150405")
	multi := len(charDirs) > 1
	failed := false
	var results []exportEntry

	for _, dir := range charDirs {
		charID := filepath.Base(dir)

		name := *metaName
		if name == "" {
			name = lookupCharName(filepath.Dir(dir), charID)
		}

		outPath := *output
		if outPath == "" {
			outPath = fmt.Sprintf("%s_macros_%s.yml", strings.ToLower(name), stamp)
		}

		if err := export.WriteFile(export.Options{CharacterDir: dir, Character: name, Scope: sc}, outPath); err != nil {
			if !p.IsJSON() {
				ew := p.Err()
				if multi {
					fmt.Fprintf(ew, "macromog export: %s: %v\n", ew.Highlight(charID), err)
				} else {
					fmt.Fprintf(ew, "macromog export: %v\n", err)
				}
			}
			results = append(results, exportEntry{Character: charID, OK: false, Error: err.Error()})
			failed = true
			continue
		}

		p.Text(func(tw *TextWriter) {
			if multi {
				fmt.Fprintf(tw, "[%s] exported macros to %s\n", tw.Highlight(charID), tw.Success(outPath))
			} else {
				fmt.Fprintf(tw, "exported macros to %s\n", tw.Success(outPath))
			}
		})
		results = append(results, exportEntry{Character: charID, Path: outPath, OK: true})
	}

	emitJSONResults(p, results, multi, failed, "export")

	if failed {
		return 1
	}
	return 0
}
