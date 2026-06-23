package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// OutputFormat controls how command output is rendered.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// Printer writes command output in the configured format.
type Printer struct {
	w      io.Writer
	format OutputFormat
}

// NewPrinter creates a Printer that writes to w in the given format.
func NewPrinter(w io.Writer, format OutputFormat) *Printer {
	return &Printer{w: w, format: format}
}

// IsJSON reports whether this printer is in JSON mode.
func (p *Printer) IsJSON() bool { return p.format == FormatJSON }

// Text calls fn with the underlying writer when in text mode.
func (p *Printer) Text(fn func(w io.Writer)) {
	if p.format != FormatJSON {
		fn(p.w)
	}
}

// JSON encodes v as indented JSON when in JSON mode.
func (p *Printer) JSON(v any) {
	if p.format == FormatJSON {
		enc := json.NewEncoder(p.w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(v)
	}
}

// extractOutputFormat scans args for a global --output flag that appears
// before the first non-flag argument (the subcommand). It strips the flag
// from the slice and returns the chosen format and the cleaned arg list.
//
// Only pre-subcommand flags are consumed so that subcommand-specific flags
// (e.g. export's --output <file>) are left untouched.
func extractOutputFormat(args []string) (OutputFormat, []string, error) {
	if len(args) == 0 {
		return FormatText, args, nil
	}

	format := FormatText
	result := make([]string, 0, len(args))
	result = append(result, args[0]) // program name

	i := 1
	for i < len(args) {
		a := args[i]
		if !strings.HasPrefix(a, "-") || a == "--" {
			// First non-flag token is the subcommand; pass everything from here unchanged.
			result = append(result, args[i:]...)
			break
		}
		if a == "--output" {
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("macromog: --output requires a value (text, json)")
			}
			val := args[i+1]
			switch val {
			case "text":
				format = FormatText
			case "json":
				format = FormatJSON
			default:
				return "", nil, fmt.Errorf("macromog: --output: unknown format %q (valid: text, json)", val)
			}
			i += 2
		} else {
			// Unknown pre-subcommand flag — keep it so the dispatcher or help handler sees it.
			result = append(result, a)
			i++
		}
	}
	return format, result, nil
}
