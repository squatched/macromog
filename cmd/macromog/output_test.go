package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestExtractOutputFormat(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantFormat OutputFormat
		wantArgs   []string
		wantErr    bool
	}{
		{
			name:       "no args",
			args:       nil,
			wantFormat: FormatText,
			wantArgs:   nil,
		},
		{
			name:       "no output flag",
			args:       []string{"macromog", "list"},
			wantFormat: FormatText,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "json before subcommand",
			args:       []string{"macromog", "--output", "json", "list"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "text before subcommand",
			args:       []string{"macromog", "--output", "text", "list"},
			wantFormat: FormatText,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "json after subcommand",
			args:       []string{"macromog", "list", "--output", "json"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "json after subcommand with other flags",
			args:       []string{"macromog", "list", "--char", "/foo", "--output", "json"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "list", "--char", "/foo"},
		},
		{
			name:       "subcommand --output file passthrough",
			args:       []string{"macromog", "export", "--output", "macros.yml"},
			wantFormat: FormatText,
			wantArgs:   []string{"macromog", "export", "--output", "macros.yml"},
		},
		{
			name:       "global json + subcommand --output file",
			args:       []string{"macromog", "--output", "json", "export", "--output", "macros.yml"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "export", "--output", "macros.yml"},
		},
		{
			name:    "unknown format before subcommand",
			args:    []string{"macromog", "--output", "xml", "list"},
			wantErr: true,
		},
		{
			name:    "missing value before subcommand",
			args:    []string{"macromog", "--output"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			format, args, err := extractOutputFormat(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (format=%q, args=%v)", format, args)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if format != tc.wantFormat {
				t.Errorf("format = %q, want %q", format, tc.wantFormat)
			}
			if len(args) != len(tc.wantArgs) {
				t.Errorf("args = %v, want %v", args, tc.wantArgs)
				return
			}
			for i := range args {
				if args[i] != tc.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tc.wantArgs[i])
				}
			}
		})
	}
}

func TestExtractOutputFormat_PostSubcommandNoValue(t *testing.T) {
	// --output with no value after subcommand should pass through unchanged.
	format, args, err := extractOutputFormat([]string{"macromog", "export", "--output"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if format != FormatText {
		t.Errorf("format = %q, want text", format)
	}
	if len(args) < 3 || args[len(args)-1] != "--output" {
		t.Errorf("--output should pass through unchanged, args = %v", args)
	}
}

func TestExtractOutputFormat_RepeatedFlag(t *testing.T) {
	// When --output appears twice before the subcommand, the last value wins.
	format, args, err := extractOutputFormat([]string{"macromog", "--output", "text", "--output", "json", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if format != FormatJSON {
		t.Errorf("format = %q, want json", format)
	}
	want := []string{"macromog", "list"}
	if len(args) != len(want) {
		t.Errorf("args = %v, want %v", args, want)
		return
	}
	for i := range args {
		if args[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestRun_OutputJSON_PostSubcommand(t *testing.T) {
	// Verify the full dispatch path: --output json placed after the subcommand.
	if got := run([]string{"macromog", "validate", "--output", "json", writeTemp(t, "version: 1\nbooks: {}\n")}); got != 0 {
		t.Errorf("run(validate --output json <valid>) = %d, want 0", got)
	}
}

// --- detectColor ---

func TestNewPrinter_NoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	p := NewPrinter(os.Stdout, FormatText)
	if p.color {
		t.Error("NewPrinter: color should be false when NO_COLOR is set")
	}
}

func TestNewPrinter_TermDumb(t *testing.T) {
	t.Setenv("TERM", "dumb")
	p := NewPrinter(os.Stdout, FormatText)
	if p.color {
		t.Error("NewPrinter: color should be false when TERM=dumb")
	}
}

func TestNewPrinter_BufferWriter_NoColor(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	if p.color {
		t.Error("NewPrinter: color should be false for non-file writer (piped/redirected output)")
	}
}

// --- TextWriter color helpers ---

func TestTextWriter_ColorOff_NoANSI(t *testing.T) {
	var buf bytes.Buffer
	tw := &TextWriter{w: &buf, color: false}
	fmt.Fprint(tw, tw.Bold("hello"))
	fmt.Fprint(tw, tw.Red("world"))
	fmt.Fprint(tw, tw.Success("ok"))
	out := buf.String()
	if strings.Contains(out, "\033[") {
		t.Errorf("color-off TextWriter must not emit ANSI codes, got: %q", out)
	}
	for _, want := range []string{"hello", "world", "ok"} {
		if !strings.Contains(out, want) {
			t.Errorf("text %q missing from output: %q", want, out)
		}
	}
}

func TestTextWriter_ColorOn_HasANSI(t *testing.T) {
	var buf bytes.Buffer
	tw := &TextWriter{w: &buf, color: true}
	fmt.Fprint(tw, tw.Bold("hello"))
	if !strings.Contains(buf.String(), "\033[") {
		t.Errorf("color-on TextWriter should emit ANSI codes, got: %q", buf.String())
	}
}

func TestTextWriter_SemanticAliases(t *testing.T) {
	tw := &TextWriter{color: true, w: &bytes.Buffer{}}
	// Semantic aliases must produce non-empty ANSI output.
	for name, got := range map[string]string{
		"Success":   tw.Success("x"),
		"Warn":      tw.Warn("x"),
		"Error":     tw.Error("x"),
		"Highlight": tw.Highlight("x"),
		"Muted":     tw.Muted("x"),
		"Label":     tw.Label("x"),
	} {
		if !strings.Contains(got, "\033[") {
			t.Errorf("%s: expected ANSI codes, got %q", name, got)
		}
		if !strings.Contains(got, "x") {
			t.Errorf("%s: text content missing from %q", name, got)
		}
	}
}

// --- ANSI sanitization ---

func TestTextWriter_StripInjectedANSI_ColorOn(t *testing.T) {
	tw := &TextWriter{color: true, w: &bytes.Buffer{}}
	injected := "\033[31mevil\033[0m" // red ANSI injected in user string
	result := tw.Cyan(injected)
	// Our cyan code should be present, injected red code should be gone.
	if strings.Contains(result, "\033[31m") {
		t.Errorf("injected red ANSI code should be stripped, got: %q", result)
	}
	if !strings.Contains(result, "evil") {
		t.Errorf("text content should be preserved, got: %q", result)
	}
	if !strings.Contains(result, "\033[36m") {
		t.Errorf("our cyan ANSI code should be present, got: %q", result)
	}
}

func TestTextWriter_StripInjectedANSI_ColorOff(t *testing.T) {
	tw := &TextWriter{color: false, w: &bytes.Buffer{}}
	injected := "\033[31mevil\033[0m"
	result := tw.Cyan(injected)
	if strings.Contains(result, "\033[") {
		t.Errorf("injected ANSI should be stripped in color-off mode, got: %q", result)
	}
	if result != "evil" {
		t.Errorf("expected bare text %q, got %q", "evil", result)
	}
}

// --- PadRight ---

func TestTextWriter_PadRight_ASCII(t *testing.T) {
	tw := &TextWriter{color: false, w: &bytes.Buffer{}}
	padded := tw.PadRight("hello", 10)
	if visibleWidth(padded) != 10 {
		t.Errorf("PadRight(ascii): visible width = %d, want 10", visibleWidth(padded))
	}
}

func TestTextWriter_PadRight_WideChars(t *testing.T) {
	tw := &TextWriter{color: false, w: &bytes.Buffer{}}
	// 玄白侍士 = 4 chars × 2 display cols = 8 display cols; pad to 28 adds 11 spaces.
	padded := tw.PadRight("玄白侍士 (aabbcc)", 28)
	if got := visibleWidth(padded); got != 28 {
		t.Errorf("PadRight(wide chars): visible width = %d, want 28", got)
	}
}

func TestTextWriter_PadRight_WideChars_WithColor(t *testing.T) {
	tw := &TextWriter{color: true, w: &bytes.Buffer{}}
	// Color codes around wide chars must not corrupt the display-width measurement.
	colored := fmt.Sprintf("%s %s", tw.Bold("玄白侍士"), tw.Muted("(aabbcc)"))
	padded := tw.PadRight(colored, 28)
	if got := visibleWidth(padded); got != 28 {
		t.Errorf("PadRight(colored wide chars): visible width = %d, want 28", got)
	}
}

func TestTextWriter_PadRight_AlreadyWide(t *testing.T) {
	tw := &TextWriter{color: false, w: &bytes.Buffer{}}
	s := "hello world this is long"
	padded := tw.PadRight(s, 5)
	if padded != s {
		t.Errorf("PadRight should not truncate strings wider than target: got %q", padded)
	}
}

// --- JSON output never carries ANSI codes ---

func TestPrinter_JSON_NoANSI_ForcedColor(t *testing.T) {
	// Even if the printer has color enabled, JSON output must be clean.
	var buf bytes.Buffer
	p := &Printer{w: &buf, format: FormatJSON, color: true}
	p.JSON(map[string]string{"key": "value"})
	if strings.Contains(buf.String(), "\033[") {
		t.Errorf("JSON output must not contain ANSI codes, got: %q", buf.String())
	}
}
