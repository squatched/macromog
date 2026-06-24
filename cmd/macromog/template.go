package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/squatched/macromog/internal/export"
	"github.com/squatched/macromog/internal/scope"
	tmpl "github.com/squatched/macromog/internal/template"
)

const templateUsage = `Usage: macromog template [flags] <output>

Generate a blank YAML template pre-structured for a given scope.
Every macro slot within scope is present with an empty name and six empty
content lines, ready to fill in.

Arguments:
  <output>              output YAML file (required)

Flags:
  --scope <selector>    scope selector (repeatable; default: full scope)
  --char-name <name>    embed character name in template

Examples:
  macromog template out.yml
  macromog template out.yml --scope B1S3
  macromog template out.yml --scope B1S3A1,C2
  macromog template out.yml --char-name Squatched
`

func runTemplate(args []string, p *Printer) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, templateUsage)
		return 0
	}

	fs := flag.NewFlagSet("template", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	charName := fs.String("char-name", "", "character name for template metadata")
	var scopeSel scopeFlags
	fs.Var(&scopeSel, "scope", "scope selector (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprint(os.Stderr, templateUsage)
		fmt.Fprintln(os.Stderr, "macromog template: output file required")
		return 1
	}
	outPath := remaining[0]

	sc, err := scope.ParseSelectors([]string(scopeSel))
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog template: invalid --scope: %v\n", err)
		return 1
	}

	doc := tmpl.Generate(sc, *charName)
	data, err := export.MarshalYAMLWithPlaceholders(doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog template: %v\n", err)
		return 1
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "macromog template: %v\n", err)
		return 1
	}

	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "template written to %s\n", tw.Success(outPath))
	})
	return 0
}
