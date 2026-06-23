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

// extractOutputFormat scans args for --output text|json and strips it from
// the arg list, returning the chosen format. The flag may appear anywhere —
// before or after the subcommand name — so that both
// "macromog --output json list" and "macromog list --output json" work.
//
// Unknown --output values are an error when seen before the subcommand (where
// there is no other command to interpret them), but are passed through
// unchanged when seen after the subcommand so that subcommand-specific
// --output flags (e.g. export's --output <file>) are left untouched.
func extractOutputFormat(args []string) (OutputFormat, []string, error) {
	if len(args) == 0 {
		return FormatText, args, nil
	}

	format := FormatText
	result := make([]string, 0, len(args))
	result = append(result, args[0]) // program name

	seenSubcommand := false
	i := 1
	for i < len(args) {
		a := args[i]
		if !seenSubcommand && (a == "--" || !strings.HasPrefix(a, "-")) {
			seenSubcommand = true
		}

		if a == "--output" {
			if i+1 >= len(args) {
				if !seenSubcommand {
					return "", nil, fmt.Errorf("macromog: --output requires a value (text, json)")
				}
				// Post-subcommand with no value: pass through for the subcommand to handle.
				result = append(result, a)
				i++
				continue
			}
			val := args[i+1]
			switch val {
			case "text":
				format = FormatText
				i += 2
			case "json":
				format = FormatJSON
				i += 2
			default:
				if !seenSubcommand {
					return "", nil, fmt.Errorf("macromog: --output: unknown format %q (valid: text, json)", val)
				}
				// Post-subcommand with non-format value (e.g. --output file.yml):
				// pass through unchanged for the subcommand to handle.
				result = append(result, a)
				i++
			}
		} else {
			result = append(result, a)
			i++
		}
	}
	return format, result, nil
}
