package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type cliState struct {
	format  string
	out     io.Writer
	printer *Printer
	code    int
}

func newRootCmd() (*cobra.Command, *cliState) {
	state := &cliState{out: os.Stdout}

	root := &cobra.Command{
		Use:           "macromog",
		Short:         "Manage FFXI macro books",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Usage()
			state.code = 1
			return nil
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch OutputFormat(state.format) {
			case FormatText, FormatJSON:
			default:
				return fmt.Errorf("--output: unknown format %q (valid: text, json)", state.format)
			}
			state.printer = NewPrinter(state.out, OutputFormat(state.format))
			return nil
		},
	}
	root.PersistentFlags().StringVar(&state.format, "output", "text", "output format: text or json")

	root.AddCommand(
		newExportCmd(state),
		newImportCmd(state),
		newTemplateCmd(state),
		newValidateCmd(state),
		newBackupCmd(state),
		newListCmd(state),
		newConfigCmd(state),
		newDebugCmd(state),
		newAgentCmd(),
	)

	return root, state
}

// execWithState runs a cobra command in isolation with a pre-set printer.
// Used by runX wrappers to preserve the test API.
func execWithState(cmd *cobra.Command, args []string, state *cliState) int {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return state.code
}
