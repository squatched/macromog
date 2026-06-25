package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
)

type exportEntry struct {
	Character string `json:"character"`
	Path      string `json:"path,omitempty"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

func (r exportEntry) ok() bool          { return r.OK }
func (r exportEntry) character() string { return r.Character }
func (r exportEntry) errMsg() string    { return r.Error }

func newExportCmd(state *cliState) *cobra.Command {
	var (
		ffxiPath    string
		installName string
		charDir     string
		charName    string
		all         bool
		dense       bool
		metaName    string
		scopeSel    []string
	)

	cmd := &cobra.Command{
		Use:   "export [path]",
		Short: "export macros from .dat files to YAML",
		Long: `Export macros from FFXI .dat files to YAML.

Without a path argument, output goes to stdout.
With a path argument:
  - Without --all: path is the output YAML file.
  - With --all: path is the directory for per-character YAML files (default: current directory).

Examples:
  macromog export
  macromog export --all
  macromog export --all ~/exports
  macromog export --dense macros.yml
  macromog export /path/to/USER/a1b2c3d4
  macromog export /path/to/USER/a1b2c3d4 macros.yml
  macromog export --char-name Squatched macros.yml
  macromog export --scope B1,5 books1and5.yml`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			sc, err := scope.ParseSelectors(scopeSel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog export: invalid --scope: %v\n", err)
				state.code = 1
				return nil
			}

			remaining := args
			if !all && charDir == "" && charName == "" && len(remaining) > 0 {
				charDir = remaining[0]
				remaining = remaining[1:]
			}

			var outPath string
			if len(remaining) > 0 {
				outPath = remaining[0]
			}

			charDirs, err := resolveCharDirs(charSelectOpts{
				charDir: charDir, charName: charName, ffxiPath: ffxiPath,
				installName: installName, all: all,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog export: %v\n", err)
				state.code = 1
				return nil
			}

			if len(charDirs) > 1 && metaName != "" {
				fmt.Fprintln(os.Stderr, "macromog export: --name requires exactly one character; use --char-dir or remove --all")
				state.code = 1
				return nil
			}

			var baseOutDir string
			if all {
				if outPath != "" {
					baseOutDir, err = filepath.Abs(outPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "macromog export: %v\n", err)
						state.code = 1
						return nil
					}
				} else {
					baseOutDir, err = os.Getwd()
					if err != nil {
						fmt.Fprintf(os.Stderr, "macromog export: %v\n", err)
						state.code = 1
						return nil
					}
				}
			}

			stamp := time.Now().UTC().Format("20060102_150405")
			multi := len(charDirs) > 1
			failed := false
			var results []exportEntry

			for _, dir := range charDirs {
				charID := filepath.Base(dir)
				name := metaName
				if name == "" {
					name = lookupCharName(filepath.Dir(dir), charID)
				}

				opts := export.Options{
					CharacterDir: dir,
					Character:    name,
					Scope:        sc,
					Dense:        dense,
				}

				var writeErr error
				var writtenPath string

				switch {
				case multi:
					writtenPath = filepath.Join(baseOutDir, fmt.Sprintf("%s_macros_%s.yml", strings.ToLower(name), stamp))
					writeErr = export.WriteFile(opts, writtenPath)
				case outPath == "":
					writeErr = export.WriteTo(state.out, opts)
				default:
					writtenPath = outPath
					writeErr = export.WriteFile(opts, writtenPath)
				}

				if writeErr != nil {
					if !p.IsJSON() {
						ew := p.Err()
						if multi {
							fmt.Fprintf(ew, "macromog export: %s: %v\n", ew.Highlight(charID), writeErr)
						} else {
							fmt.Fprintf(ew, "macromog export: %v\n", writeErr)
						}
					}
					results = append(results, exportEntry{Character: charID, OK: false, Error: writeErr.Error()})
					failed = true
					continue
				}

				p.Text(func(tw *TextWriter) {
					if writtenPath == "" {
						return
					}
					if multi {
						fmt.Fprintf(tw, "[%s] exported macros to %s\n", tw.Highlight(charID), tw.Success(writtenPath))
					} else {
						fmt.Fprintf(tw, "exported macros to %s\n", tw.Success(writtenPath))
					}
				})
				results = append(results, exportEntry{Character: charID, Path: writtenPath, OK: true})
			}

			emitJSONResults(p, results, multi, failed, "export")
			if failed {
				state.code = 1
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ffxiPath, "ffxi-path", "", "FFXI install root")
	cmd.Flags().StringVar(&installName, "install", "", "named FFXI install from config")
	cmd.Flags().StringVar(&charDir, "char-dir", "", "character USER directory")
	cmd.Flags().StringVar(&charName, "char-name", "", "friendly character name from config")
	cmd.Flags().BoolVar(&all, "all", false, "export all discovered characters without prompting")
	cmd.Flags().BoolVar(&dense, "dense", false, "include all in-scope macro slots even if empty")
	cmd.Flags().StringVar(&metaName, "name", "", "character name for YAML metadata")
	cmd.Flags().StringArrayVar(&scopeSel, "scope", nil, "scope selector (repeatable; e.g. B1S3, B1,5)")

	return cmd
}

func runExport(args []string, p *Printer) int {
	state := &cliState{printer: p, out: os.Stdout}
	return execWithState(newExportCmd(state), args, state)
}
