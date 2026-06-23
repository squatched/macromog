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

type validateResult struct {
	File   string   `json:"file"`
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

func runValidate(args []string, p *Printer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		if len(args) == 0 {
			fmt.Fprint(os.Stderr, validateUsage)
			return 1
		}
		fmt.Fprint(os.Stdout, validateUsage)
		return 0
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog validate: %v\n", err)
		return 1
	}

	errs := validate.Validate(data)

	if len(errs) == 0 {
		p.Text(func(tw *TextWriter) {
			fmt.Fprintf(tw, "%s: %s\n", tw.Highlight(filename), tw.Success("OK"))
		})
		p.JSON(validateResult{File: filename, Valid: true})
		return 0
	}

	errStrs := make([]string, len(errs))
	for i, e := range errs {
		errStrs[i] = e.Error()
	}

	// In text mode print details to stderr; in JSON mode the errors array
	// in the output carries them so stderr stays clean.
	if !p.IsJSON() {
		ew := p.Err()
		fmt.Fprintf(ew, "%s: %d validation error(s):\n", ew.Highlight(filename), len(errs))
		for _, e := range errs {
			fmt.Fprintf(ew, "  %s\n", ew.Error(e.Error()))
		}
	}
	p.JSON(validateResult{File: filename, Valid: false, Errors: errStrs})
	return 1
}
