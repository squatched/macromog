package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/validate"
)

type validateResult struct {
	File   string   `json:"file"`
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

func newValidateCmd(state *cliState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "validate a YAML file against the schema",
		Long: `Validate a YAML macro file against the macromog schema.
Exits 0 if valid, 1 if errors are found.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			filename := args[0]
			data, err := os.ReadFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog validate: %v\n", err)
				state.code = 1
				return nil
			}

			errs := validate.Validate(data)
			if len(errs) == 0 {
				p.Text(func(tw *TextWriter) {
					fmt.Fprintf(tw, "%s: %s\n", tw.Highlight(filename), tw.Success("OK"))
				})
				p.JSON(validateResult{File: filename, Valid: true})
				return nil
			}

			errStrs := make([]string, len(errs))
			for i, e := range errs {
				errStrs[i] = e.Error()
			}
			if !p.IsJSON() {
				ew := p.Err()
				fmt.Fprintf(ew, "%s: %d validation error(s):\n", ew.Highlight(filename), len(errs))
				for _, e := range errs {
					fmt.Fprintf(ew, "  %s\n", ew.Error(e.Error()))
				}
			}
			p.JSON(validateResult{File: filename, Valid: false, Errors: errStrs})
			state.code = 1
			return nil
		},
	}
	return cmd
}

func runValidate(args []string, p *Printer) int {
	state := &cliState{printer: p, out: os.Stdout}
	return execWithState(newValidateCmd(state), args, state)
}
