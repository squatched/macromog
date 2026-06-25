package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/importer"
	"github.com/squatched/macromog/internal/scope"
	"gopkg.in/yaml.v3"
)

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

func (r importEntry) ok() bool          { return r.OK }
func (r importEntry) character() string { return r.Character }
func (r importEntry) errMsg() string    { return r.Error }

type importSetInfo struct {
	FileName string `json:"file"`
	Book     int    `json:"book"`
	Set      int    `json:"set"`
}

func newImportCmd(state *cliState) *cobra.Command {
	var (
		chars    charSelectOpts
		noBackup bool
		dryRun   bool
		scopeSel []string
	)

	cmd := &cobra.Command{
		Use:   "import <file> [<char-dir>]",
		Short: "import macros from YAML into .dat files",
		Long: `Import macros from YAML into FFXI .dat files.
A timestamped backup of current macros is created before writing (use
--no-backup to skip).

Examples:
  macromog import mymacros.yml
  macromog import mymacros.yml /path/to/USER/a1b2c3d4
  macromog import --all mymacros.yml
  macromog import --char-dir /path/to/USER/a1b2c3d4 mymacros.yml
  macromog import --char-name Squatched mymacros.yml
  macromog import --dry-run mymacros.yml /path/to/USER/a1b2c3d4
  macromog import --scope B1 mymacros.yml`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			yamlPath := args[0]
			if chars.charDir == "" && chars.charName == "" && len(args) > 1 {
				chars.charDir = args[1]
			}

			var importScope scope.Scope
			if len(scopeSel) > 0 {
				var err error
				importScope, err = scope.ParseSelectors(scopeSel)
				if err != nil {
					fmt.Fprintf(os.Stderr, "macromog import: invalid --scope: %v\n", err)
					state.code = 1
					return nil
				}
			}

			charDirs, err := resolveCharDirs(chars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
				state.code = 1
				return nil
			}

			yamlAbs, err := filepath.Abs(yamlPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
				state.code = 1
				return nil
			}

			if !importScope.IsZero() && !dryRun {
				if confirmed, err := confirmScopeOverride(yamlAbs, importScope, stdinReader()); err != nil {
					fmt.Fprintf(os.Stderr, "macromog import: %v\n", err)
					state.code = 1
					return nil
				} else if !confirmed {
					fmt.Fprintln(os.Stderr, "macromog import: aborted")
					state.code = 1
					return nil
				}
			}

			multi := len(charDirs) > 1
			failed := false
			var results []importEntry

			for _, dir := range charDirs {
				charID := filepath.Base(dir)
				result, ierr := importer.Import(importer.Options{
					CharacterDir: dir,
					YAMLPath:     yamlAbs,
					Scope:        importScope,
					Backup:       !noBackup,
					DryRun:       dryRun,
				})
				if ierr != nil {
					if !p.IsJSON() {
						ew := p.Err()
						if multi {
							fmt.Fprintf(ew, "macromog import: %s: %v\n", ew.Highlight(charID), ierr)
						} else {
							fmt.Fprintf(ew, "macromog import: %v\n", ierr)
						}
					}
					results = append(results, importEntry{
						Character: charID,
						YAMLFile:  filepath.Base(yamlPath),
						DryRun:    dryRun,
						OK:        false,
						Error:     ierr.Error(),
					})
					failed = true
					continue
				}

				p.Text(func(tw *TextWriter) {
					if dryRun {
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
					DryRun:    dryRun,
					OK:        true,
				}
				if result.BackupDir != "" {
					entry.BackupPath = result.BackupDir
				}
				if dryRun {
					sets := make([]importSetInfo, len(result.Sets))
					for i, s := range result.Sets {
						sets[i] = importSetInfo{FileName: s.FileName, Book: s.Book, Set: s.Set}
					}
					entry.DryRunSets = sets
				}
				results = append(results, entry)
			}

			emitJSONResults(p, results, multi, failed, "import")
			if failed {
				state.code = 1
			}
			return nil
		},
	}

	addCharFlags(cmd, &chars)
	cmd.Flags().BoolVar(&chars.all, "all", false, "import into all discovered characters without prompting")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip automatic backup before writing")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "validate and show what would be written, without writing")
	cmd.Flags().StringArrayVar(&scopeSel, "scope", nil, "override import scope (repeatable; e.g. B1, B1S3)")

	return cmd
}

// confirmScopeOverride reads the YAML scope field and, if the import scope
// exceeds it, asks the user to confirm. Returns true if import should proceed.
// r is read for the user's response; pass os.Stdin in production.
func confirmScopeOverride(yamlPath string, importScope scope.Scope, r io.Reader) (bool, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return false, err
	}
	var doc export.Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, err
	}
	if !importScope.Exceeds(doc.Scope) {
		return true, nil
	}

	yamlLevel := doc.Scope.Level
	if doc.Scope.IsZero() {
		yamlLevel = "none (legacy file)"
	}
	fmt.Fprintf(os.Stderr,
		"Warning: --scope %s exceeds the YAML's embedded scope (%s).\n"+
			"Entries within the expanded scope but absent from the YAML will be cleared.\n"+
			"Proceed? [y/N] ",
		importScope.Level, yamlLevel,
	)

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false, nil
	}
	return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y", nil
}

func runImport(args []string, p *Printer) int {
	state := &cliState{printer: p, out: os.Stdout}
	return execWithState(newImportCmd(state), args, state)
}
