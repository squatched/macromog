package main

import (
	"fmt"
	"os"
)

const usage = `Usage: macromog [--output <format>] <command> [flags]

Commands:
  alias     assign a friendly name to a character folder
  export    export macros from .dat files to YAML
  import    import macros from YAML into .dat files (auto-backups first)
  validate  validate a YAML file against the schema
  backup    create a timestamped backup of all macro .dat files
  list      list detected characters and macro books

Global flags:
  --output <format>   output format: text (default) or json
  --ffxi-path <path>  path to FFXI install (auto-detected if possible)
  --char-dir <id>     character folder (hex ID or path)
  --char-name <name>  character alias (set with 'macromog alias')

Run 'macromog <command> --help' for command-specific flags.
`

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	format, args, err := extractOutputFormat(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}

	p := NewPrinter(os.Stdout, format)

	switch args[1] {
	case "alias":
		return runAlias(args[2:], p)
	case "export":
		return runExport(args[2:], p)
	case "import":
		return runImport(args[2:], p)
	case "validate":
		return runValidate(args[2:], p)
	case "backup":
		return runBackup(args[2:], p)
	case "list":
		return runList(args[2:], p)
	case "--help", "-h", "help":
		fmt.Fprint(os.Stdout, usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "macromog: unknown command %q\n\n%s", args[1], usage)
		return 1
	}
}

func unimplemented(cmd string) int {
	fmt.Fprintf(os.Stderr, "macromog %s: not yet implemented\n", cmd)
	return 1
}
