package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
	tmpl "github.com/squatched/macromog/internal/template"
)

func newTemplateCmd(state *cliState) *cobra.Command {
	var (
		charName string
		scopeSel []string
	)

	cmd := &cobra.Command{
		Use:   "template [output]",
		Short: "generate a blank YAML template for a given scope",
		Long: `Generate a blank YAML template pre-structured for a given scope.
Every macro slot within scope is present with an empty name and six empty
content lines, ready to fill in.

Without an output argument, the template is written to stdout.

Examples:
  macromog template
  macromog template out.yml
  macromog template --scope B1S3 out.yml
  macromog template out.yml --scope B1S3A1,C2
  macromog template --char-name Squatched out.yml`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			sc, err := scope.ParseSelectors(scopeSel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog template: invalid --scope: %v\n", err)
				state.code = 1
				return nil
			}

			doc := tmpl.Generate(sc, charName)
			data, err := export.MarshalYAMLWithPlaceholders(doc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog template: %v\n", err)
				state.code = 1
				return nil
			}

			if len(args) == 0 {
				if _, err := state.out.Write(data); err != nil {
					fmt.Fprintf(os.Stderr, "macromog template: %v\n", err)
					state.code = 1
				}
				return nil
			}

			outPath := args[0]
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "macromog template: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				fmt.Fprintf(tw, "template written to %s\n", tw.Success(outPath))
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&charName, "char-name", "", "embed character name in template")
	cmd.Flags().StringArrayVar(&scopeSel, "scope", nil, "scope selector (repeatable; e.g. B1S3, B1S3A1,C2)")

	return cmd
}

func runTemplate(args []string, p *Printer) int {
	state := &cliState{printer: p, out: os.Stdout}
	return execWithState(newTemplateCmd(state), args, state)
}
