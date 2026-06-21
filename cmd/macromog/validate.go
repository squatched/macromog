package main

import (
	"fmt"
	"os"

	"github.com/squatched/macromog/internal/validate"
)

const validateUsage = `Usage: macromog validate <file>

Validate a YAML macro file against the macromog schema.
Exits 0 if valid, 1 if errors are found.
`

func runValidate(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		if len(args) == 0 {
			fmt.Fprint(os.Stderr, validateUsage)
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, validateUsage)
		return
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog validate: %v\n", err)
		os.Exit(1)
	}

	errs := validate.Validate(data)
	if len(errs) == 0 {
		fmt.Printf("%s: OK\n", filename)
		return
	}

	fmt.Fprintf(os.Stderr, "%s: %d validation error(s):\n", filename, len(errs))
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  %s\n", e)
	}
	os.Exit(1)
}
