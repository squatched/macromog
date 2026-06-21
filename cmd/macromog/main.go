package main

import (
	"fmt"
	"os"
)

const usage = `Usage: macromog <command> [flags]

Commands:
  export    export macros from .dat files to YAML
  import    import macros from YAML into .dat files (auto-backups first)
  validate  validate a YAML file against the schema
  backup    create a timestamped backup of all macro .dat files
  list      list detected characters and macro books

Global flags:
  --ffxi-path <path>  path to FFXI install (auto-detected if possible)
  --char <id>         character folder (hex ID or path)

Run 'macromog <command> --help' for command-specific flags.
`

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}

	switch args[1] {
	case "export":
		return runExport(args[2:])
	case "import":
		return unimplemented("import")
	case "validate":
		return runValidate(args[2:])
	case "backup":
		return unimplemented("backup")
	case "list":
		return unimplemented("list")
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
