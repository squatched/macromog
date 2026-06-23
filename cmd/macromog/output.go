package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
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
	color  bool
}

// NewPrinter creates a Printer that writes to w in the given format.
// Color is enabled automatically when w is a terminal and NO_COLOR is not set.
func NewPrinter(w io.Writer, format OutputFormat) *Printer {
	colorOn := detectColor(w)
	actual := w
	if colorOn {
		if f, ok := w.(*os.File); ok {
			actual = colorable.NewColorable(f)
		}
	}
	return &Printer{w: actual, format: format, color: colorOn}
}

func detectColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// IsJSON reports whether this printer is in JSON mode.
func (p *Printer) IsJSON() bool { return p.format == FormatJSON }

// Err returns a TextWriter targeting stderr, with color auto-detected for
// stderr independently of stdout. Use it for error messages in text mode.
func (p *Printer) Err() *TextWriter {
	errColor := detectColor(os.Stderr)
	w := io.Writer(os.Stderr)
	if errColor {
		w = colorable.NewColorableStderr()
	}
	return &TextWriter{w: w, color: errColor}
}

// Text calls fn with a TextWriter when in text mode.
func (p *Printer) Text(fn func(tw *TextWriter)) {
	if p.format != FormatJSON {
		fn(&TextWriter{w: p.w, color: p.color})
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

// TextWriter is passed to Text callbacks. It implements io.Writer directly
// (for use with fmt.Fprintf) and provides inline color span helpers.
type TextWriter struct {
	w     io.Writer
	color bool
}

// Write implements io.Writer so fmt.Fprintf(tw, ...) works.
func (tw *TextWriter) Write(p []byte) (n int, err error) {
	return tw.w.Write(p)
}

const ansiReset = "\033[0m"

func (tw *TextWriter) wrap(code, s string) string {
	// Strip any pre-existing ANSI codes from user-supplied strings so they
	// cannot inject sequences into the output or corrupt our own wrapping.
	clean := ansiEscape.ReplaceAllString(s, "")
	if !tw.color {
		return clean
	}
	return code + clean + ansiReset
}

// Raw color span helpers — return a styled string for inline embedding.
func (tw *TextWriter) Bold(s string) string    { return tw.wrap("\033[1m", s) }
func (tw *TextWriter) Dim(s string) string     { return tw.wrap("\033[2m", s) }
func (tw *TextWriter) Red(s string) string     { return tw.wrap("\033[31m", s) }
func (tw *TextWriter) Green(s string) string   { return tw.wrap("\033[32m", s) }
func (tw *TextWriter) Yellow(s string) string  { return tw.wrap("\033[33m", s) }
func (tw *TextWriter) Blue(s string) string    { return tw.wrap("\033[34m", s) }
func (tw *TextWriter) Magenta(s string) string { return tw.wrap("\033[35m", s) }
func (tw *TextWriter) Cyan(s string) string    { return tw.wrap("\033[36m", s) }
func (tw *TextWriter) White(s string) string   { return tw.wrap("\033[37m", s) }

// Semantic aliases.
func (tw *TextWriter) Success(s string) string   { return tw.Green(s) }
func (tw *TextWriter) Warn(s string) string      { return tw.Yellow(s) }
func (tw *TextWriter) Error(s string) string     { return tw.Red(s) }
func (tw *TextWriter) Highlight(s string) string { return tw.Cyan(s) }
func (tw *TextWriter) Muted(s string) string     { return tw.Dim(s) }
func (tw *TextWriter) Label(s string) string     { return tw.Bold(s) }

// PadRight pads s to the given visible width, ignoring embedded ANSI codes.
// Use this instead of %-Ns in format strings when s may contain color codes.
func (tw *TextWriter) PadRight(s string, width int) string {
	vis := visibleWidth(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

var ansiEscape = regexp.MustCompile(`\033\[[0-9;]*m`)

func visibleWidth(s string) int {
	return runewidth.StringWidth(ansiEscape.ReplaceAllString(s, ""))
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
