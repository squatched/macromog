package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/backup"
)

type backupEntry struct {
	Character string `json:"character"`
	Path      string `json:"path,omitempty"`
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

func (r backupEntry) ok() bool          { return r.OK }
func (r backupEntry) character() string { return r.Character }
func (r backupEntry) errMsg() string    { return r.Error }

func newBackupCmd(state *cliState) *cobra.Command {
	var (
		chars   charSelectOpts
		outDir  string
		inPlace bool
	)

	cmd := &cobra.Command{
		Use:   "backup [<char-dir>]",
		Short: "create a timestamped backup of all macro .dat files",
		Long: `Create a timestamped backup of all macro .dat files for a character.
The backup directory is named <char-id>_YYYYMMDD_HHMMSS.

Examples:
  macromog backup
  macromog backup /path/to/USER/a1b2c3d4
  macromog backup --all --in-place
  macromog backup --char-dir /path/to/USER/a1b2c3d4 --out ~/macro-backups
  macromog backup --char-name Squatched --out ~/macro-backups`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			if chars.charDir == "" && chars.charName == "" && len(args) > 0 {
				chars.charDir = args[0]
			}

			if outDir != "" && inPlace {
				fmt.Fprintln(os.Stderr, "macromog backup: --out and --in-place are mutually exclusive")
				state.code = 1
				return nil
			}

			charDirs, err := resolveCharDirs(chars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
				state.code = 1
				return nil
			}

			var baseDestDir string
			if !inPlace {
				if outDir != "" {
					baseDestDir, err = filepath.Abs(outDir)
					if err != nil {
						fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
						state.code = 1
						return nil
					}
				} else {
					baseDestDir, err = os.Getwd()
					if err != nil {
						fmt.Fprintf(os.Stderr, "macromog backup: %v\n", err)
						state.code = 1
						return nil
					}
				}
			}

			multi := len(charDirs) > 1
			failed := false
			var results []backupEntry

			for _, dir := range charDirs {
				charID := filepath.Base(dir)
				destDir := baseDestDir
				if inPlace {
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
				state.code = 1
			}
			return nil
		},
	}

	addCharFlags(cmd, &chars)
	cmd.Flags().BoolVar(&chars.all, "all", false, "back up all discovered characters")
	cmd.Flags().StringVar(&outDir, "out", "", "directory to write the backup into (default: current directory)")
	cmd.Flags().BoolVar(&inPlace, "in-place", false, "write backup into <char-dir>/backups/")

	return cmd
}

func runBackup(args []string, p *Printer) int {
	state := &cliState{printer: p, out: os.Stdout}
	return execWithState(newBackupCmd(state), args, state)
}
