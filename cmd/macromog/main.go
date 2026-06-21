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
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "export":
		unimplemented("export")
	case "import":
		unimplemented("import")
	case "validate":
		unimplemented("validate")
	case "backup":
		unimplemented("backup")
	case "list":
		unimplemented("list")
	case "--help", "-h", "help":
		fmt.Fprint(os.Stdout, usage)
	default:
		fmt.Fprintf(os.Stderr, "macromog: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}

func unimplemented(cmd string) {
	fmt.Fprintf(os.Stderr, "macromog %s: not yet implemented\n", cmd)
	os.Exit(1)
}
